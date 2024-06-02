package bridge

import (
	"github.com/Nevermore12321/dockergsh/internal/engine"
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
		network        *net.IPNet
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

	// 查看网桥是否已经存在
	networkdriver.GetIfaceAddr()

	return 0
}
