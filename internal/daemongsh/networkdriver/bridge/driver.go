package bridge

import (
	"github.com/Nevermore12321/dockergsh/internal/daemongsh/networkdriver"
	"github.com/Nevermore12321/dockergsh/internal/engine"
	"github.com/Nevermore12321/dockergsh/pkg/networkfs/resolvconf"
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
	nameserver := []string{} // /etc/resolve.conf
	resolvConf, _ := resolvconf.Get()
	// 这里不检查错误，因为并不真正关心是否无法读取 resolv.conf。
	// 因此，如果 resolvConf 为 nil，会跳过追加。
	if resolvConf != nil {
		nameserver = append(nameserver, resolvconf.GetNameserversAsCIDR(resolvConf)...)
	}
}

// CreateBridgeIface 在主机系统上创建一个名为“ifaceName”的网桥接口，并尝试使用不与主机上任何其他接口冲突的地址来配置它。如果找不到不冲突的地址，则会返回错误。
