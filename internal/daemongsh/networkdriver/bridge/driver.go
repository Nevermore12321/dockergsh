package bridge

import (
	"fmt"
	"github.com/Nevermore12321/dockergsh/external/libcontainer/netlink"
	"github.com/Nevermore12321/dockergsh/internal/daemongsh/networkdriver"
	"github.com/Nevermore12321/dockergsh/internal/engine"
	"github.com/Nevermore12321/dockergsh/pkg/networkfs/resolvconf"
	"github.com/Nevermore12321/dockergsh/pkg/parse/kernel"
	log "github.com/sirupsen/logrus"
	"net"
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
	if err := netlink.NetworkLinkAddIp(iface, ipNet, ipAddr); err != nil {
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
