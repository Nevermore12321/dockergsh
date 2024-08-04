package bridge

import (
	"fmt"
	"github.com/Nevermore12321/dockergsh/external/libcontainer/netlink"
	"github.com/Nevermore12321/dockergsh/internal/daemongsh/networkdriver"
	"github.com/Nevermore12321/dockergsh/internal/engine"
	"github.com/Nevermore12321/dockergsh/pkg/iptables"
	"github.com/Nevermore12321/dockergsh/pkg/networkfs/resolvconf"
	"github.com/Nevermore12321/dockergsh/pkg/parse/kernel"
	log "github.com/sirupsen/logrus"
	"net"
	"os"
)

const (
	DefaultNetworkBridge = "dockergsh0"
)

var (
	// 设置网桥的名称
	bridgeIface string

	// dockergsh -p 端口映射时，默认的宿主机的绑定 ip，默认 0.0.0.0
	defaultBindingIp = net.ParseIP("0.0.0.0")

	// 可选的默认网关 ip 地址 cidr
	addrs = []string{
		"172.17.42.1/16",
		"10.0.42.1/16",
		"10.1.42.1/16",
		"10.42.42.1/16",
		"172.16.42.1/24",
		"172.16.43.1/24",
		"172.16.44.1/24",
		"10.0.42.1/24",
		"10.0.43.1/24",
		"192.168.42.1/24",
		"192.168.43.1/24",
		"192.168.44.1/24",
	}
)

// InitDriver builtins 注册的 init_networkdriver job 的handler处理函数
// 1. 获取为Docker服务的网络设备地址。
// 2. 创建指定IP地址的网桥(名为 dockergsh0)。
// 3. 配置网络iptables规则。
// 4. 另外还为 eng 对象注册了多个 Handler，如 allocate_interface、release_interface、allocate_port以及link等。
func InitDriver(job *engine.Job) engine.Status {
	var (
		network        *net.IPNet                                      // 网桥信息结构体
		enableIptables = job.GetEnvBool("EnableIptables")              // 是否开启 iptables
		icc            = job.GetEnvBool("InterContainerCommunication") // 是否允许容器相互通信
		ipForward      = job.GetEnvBool("EnableIpForward")             // 是否允许 ip forward
		bridgeIp       = job.GetEnv("BridgeIp")                        // 默认网桥的 ip 地址
	)

	// 如果指定了修改绑定的默认 ip，修改
	if defaultIp := job.GetEnv("DefaultBindingIp"); defaultIp != "" {
		defaultBindingIp = net.ParseIP(defaultIp)
	}

	// 如果设置了网桥名称，修改，否则使用默认网桥名称dockergsh0
	bridgeIface = job.GetEnv("BridgeIface")
	useDefaultBridge := false
	if bridgeIface == "" {
		useDefaultBridge = true
		bridgeIface = DefaultNetworkBridge
	}

	// 判断是否创建dockergsh0网桥，若Docker专属网桥已存在，则继续往下执行；否则，创建docker0网桥
	addr, err := networkdriver.GetIfaceAddr(bridgeIface)
	if err != nil { // 获取网桥信息失败
		// 如果不使用默认网桥，是用户指定的网桥，那么不去创建，直接报错
		if !useDefaultBridge {
			job.Logf("bridge not found: %s", bridgeIface)
			return job.Error(err)
		}

		// 如果使用的是默认网桥，并且没有找到，则创建
		job.Logf("creating a new bridge for %s", bridgeIface)
		err := createBridge(bridgeIp)
		if err != nil {
			return job.Error(err)
		}

		// 创建完成后，再去重新获取网桥信息
		addr, err = networkdriver.GetIfaceAddr(bridgeIface)
		if err != nil {
			return job.Error(err)
		}
		network = addr.(*net.IPNet)
	} else { // 获取网桥信息成功
		network = addr.(*net.IPNet)
		// 验证网桥 ip 是否与 BridgeIP 指定的 ip 匹配
		if bridgeIp != "" {
			// ParseCIDR(192.0.2.1/24) => IP：192.0.2.1 and network 192.0.2.0/24
			bip, _, err := net.ParseCIDR(bridgeIp)
			if err != nil {
				return job.Error(err)
			}
			if !network.IP.Equal(bip) {
				return job.Errorf("bridge ip (%s) does not match existing bridge configuration %s", network.IP, bip)
			}
		}
	}

	// 创建完网桥之后，Docker Daemon为Docker容器以及宿主机配置iptables规则
	// 为Docker容器之间的link操作提供iptables防火墙支持
	if enableIptables {
		if err := setupIptables(addr, icc); err != nil {
			return job.Error(err)
		}
	}

	// 启用系统数据包转发功能，将 /proc/sys/net/ipv4/ip_forward 置 1
	if ipForward {
		// enable ipv4 forwarding
		if err := os.WriteFile("/proc/sys/net/ipv4/ip_forward", []byte{'1', '\n'}, 0644); err != nil {
			job.Logf("WARNING: unable to enable IPv4 forwarding: %s\n", err)
		}

	}

	// 创建DOCKERGSH链，在创建Docker容器时实现容器与宿主机的端口映射
	// 每次启动前，先把之前的 iptables 清除
	if err := iptables.RemoveExistingChain("DOCKERGSH"); err != nil {
		return job.Error(err)
	}
	// 创建新 chain
	if enableIptables {
		chain, err := iptables.NewChain("DOCKERGSH", bridgeIface)
		if err != nil {
			return job.Error(err)
		}
		// todo 添加 端口映射的 iptables 规则
		portmapper.SetIptablesChain(chain)
	}

	return 0
}

// 在宿主机上创建指定名称网桥设备，并为该网桥设备配置一个与其他设备不冲突的网络地址。
func createBridge(bridgeIp string) error {
	nameservers := []string{} // /etc/resolve.conf
	resolvConf, _ := resolvconf.Get()
	// 这里不检查错误，因为并不真正关心是否无法读取 resolv.conf。
	// 因此，如果 resolvConf 为 nil，会跳过追加。
	if resolvConf != nil {
		nameservers = append(nameservers, resolvconf.GetNameserversAsCIDR(resolvConf)...)
	}

	var ifaceAddr string    // bridgeIp
	if len(bridgeIp) != 0 { // 如果用户指定了 bridgeIp ，则校验
		_, _, err := net.ParseCIDR(bridgeIp) // 校验 ip 格式
		if err != nil {
			return err
		}
		ifaceAddr = bridgeIface
	} else { // 如果用户没有指定 bridgeIp，从默认的 ip cidr 中选一个（并且检查与 nameserver 没有网络堆叠）
		for _, addr := range addrs { // 遍历默认网关 ip
			_, dockergshNetwork, err := net.ParseCIDR(addr) // 校验
			if err != nil {
				return err
			}

			// 如果当前 cidr 没有与 nameservers 有 ip 堆叠，那么就选中作为 bridgeIP
			if err := networkdriver.CheckNameserverOverlaps(nameservers, dockergshNetwork); err == nil {
				ifaceAddr = addr
				break
			} else { // 如果有堆叠，打印日志后，继续检查下一个
				log.Debugf("%s %s", addr, err)
			}
		}
	}

	// 如果所有默认网关都检查失败，退出
	if ifaceAddr == "" {
		return fmt.Errorf("could not find a free IP address range for interface '%s'. Please configure its address manually and run 'docker -b %s'", bridgeIface, bridgeIface)
	}

	log.Debugf("Creating bridge %s with network %s", bridgeIface, ifaceAddr)

	// 创建网桥，注意，这里只添加了 mac 地址，还没有设置 ip 地址
	if err := createBridgeIface(bridgeIface); err != nil {
		return err
	}

	// 创建完成后，检查是否存在
	iface, err := net.InterfaceByName(bridgeIface)
	if err != nil {
		return err
	}

	// 解析网桥 ip 地址
	ipAddr, ipNet, err := net.ParseCIDR(ifaceAddr)
	if err != nil {
		return err
	}

	// 给网桥设置 ip 地址
	if err := netlink.NetworkLinkAddIp(iface, ipAddr, ipNet); err != nil {
		return fmt.Errorf("unable to add private network: %s", err)
	}

	// 启动网桥设备
	if err := netlink.NetworkLinkUp(iface); err != nil {
		return fmt.Errorf("unable to start network bridge: %s", err)
	}

	return nil
}

// CreateBridgeIface 在主机系统上创建一个名为“ifaceName”的网桥接口，并尝试设置 mac 地址
func createBridgeIface(name string) error {
	// 获取 kernel 版本
	kernelVersion, err := kernel.GetKernelVersion()
	// 内核版本 > 3.3 时才支持设置网桥的 MAC 地址，其他都不设置 mac
	setBridgeMacAddr := err == nil && (kernelVersion.Kernel >= 3 && kernelVersion.Major >= 3)

	log.Debugf("setting bridge mac address = %v", setBridgeMacAddr)

	return netlink.CreateBridge(name, setBridgeMacAddr)
}

// 设置 iptables 规则
// addr - 地址为Docker网桥的网络地址
// icc - 为true，即允许Docker容器间互相访问
func setupIptables(addr net.Addr, icc bool) error {
	// 1. 使用iptables工具开启新建网桥的NAT功能
	// 全部命令为： iptables -I POSTROUTING -t nat -s [dockergsh0_ip] ！ -o dockergsh0 -j MASQUERADE
	natArgs := []string{"POSTROUTING", "-t", "nat", "-s", addr.String(), "!", "-o", bridgeIface, "-j", "MASQUERADE"}
	if !iptables.Exists(natArgs...) { // 如果iptables规则不存在，则添加
		output, err := iptables.Raw(append([]string{"-I"}, natArgs...)...)
		if err != nil {
			return fmt.Errorf("unable to enable network bridge NAT: %s", err)
		} else if len(output) != 0 {
			return fmt.Errorf("error iptables postrouting: %s", output)
		}
	}

	// 2. 通过icc参数，决定是否允许Docker容器间的通信，并制定相应iptables的Forward链
	// 即：iptables -I FORWARD -i dockergsh0 -o dockergsh0 -j ACCEPT
	var (
		args       = []string{"FORWARD", "-i", bridgeIface, "-o", "output", "-j"} // FORWARD 链，-j 后根 accept 还是 drop
		acceptArgs = append(args, "ACCEPT")
		dropArgs   = append(args, "DROP")
	)

	if !icc { // 不允许容器间通信
		// 先删除 ACCEPT 规则
		iptables.Raw(append([]string{"-D"}, acceptArgs...)...)

		// 添加 drop 规则
		if !iptables.Exists(dropArgs...) {
			log.Debugf("Disable inter-container communication")
			if output, err := iptables.Raw(append([]string{"-I"}, dropArgs...)...); err != nil {
				return fmt.Errorf("unable to prevent intercontainer communication: %s", err)
			} else if len(output) != 0 {
				return fmt.Errorf("error disabling intercontainer communication: %s", output)
			}
		}
	} else { // 允许容器间通信
		// 县删除 drop 规则
		iptables.Raw(append([]string{"-D"}, dropArgs...)...)

		// 添加 accept 规则
		if !iptables.Exists(acceptArgs...) {
			log.Debugf("Enable inter-container communication")
			if output, err := iptables.Raw(append([]string{"-I"}, acceptArgs...)...); err != nil {
				return fmt.Errorf("unable to allow intercontainer communication: %s", err)
			} else if len(output) != 0 {
				return fmt.Errorf("error enabling intercontainer communication: %s", output)
			}
		}
	}

	// 3. 允许接受从容器发出，且目标地址不是容器的数据包，也就是允许所有从docker0发出且不是继续发向docker0的数据包
	// 即命令：iptables -I FORWARD -i docker0 ! -o docker0 -j ACCEPT
	outgoingArgs := []string{"FORWARD", "-i", bridgeIface, "!", "-o", bridgeIface, "-j", "ACCEPT"}
	if !iptables.Exists(outgoingArgs...) {
		if output, err := iptables.Raw(append([]string{"-I"}, outgoingArgs...)...); err != nil {
			return fmt.Errorf("unable to allow outgoing packets: %s", err)
		} else if len(output) != 0 {
			return fmt.Errorf("error iptables allow outgoing: %s", output)
		}
	}

	// 4. 允许接受从容器发出，且目标地址不是容器的数据包。换言之，允许所有从docker0发出且不是继续发向docker0的数据包
	// 即命令：iptables -I FORWARD -o docker0 -m conntrack --ctstate RELATED,ESTABLISHED -j ACCEPT
	existingArgs := []string{"FORWARD", "-o", bridgeIface, "-m", "conntrack", "--ctstate", "RELATED,ESTABLISHED", "-j", "ACCEPT"}
	if !iptables.Exists(existingArgs...) {
		if output, err := iptables.Raw(append([]string{"-I"}, existingArgs...)...); err != nil {
			return fmt.Errorf("unable to allow incoming packets: %s", err)
		} else if len(output) != 0 {
			return fmt.Errorf("error iptables allow incoming: %s", output)
		}
	}
	return nil
}
