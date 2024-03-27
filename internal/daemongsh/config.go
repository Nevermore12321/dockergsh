package daemongsh

import (
	"github.com/Nevermore12321/dockergsh/internal/daemongsh/networkdriver"
	"github.com/Nevermore12321/dockergsh/internal/utils"
	"github.com/urfave/cli/v2"
	"net"
)

// network 默认配置
var (
	DefaultNetworkMtu    = 1500 // mtu 默认值
	DisableNetworkBridge = "none"
)

// Config 在启动 dockergsh 守护进程时传递给它的配置设置，例如：`docker -d -e lxc`
type Config struct {
	PidFile                     string              // daemon 所属进程的 PID 文件
	Root                        string              // dockergsh 运行时所使用的 root 路径
	AutoRestart                 bool                // 是否支持创建的容器自动重启
	Dns                         []string            // daemon 创建容器的默认 DNS server 地址
	DnsSearch                   []string            // dockergsh 使用指定的 DNS 查找地址
	EnableIptables              bool                // 是否启用 daemon 的 iptables 功能
	EnableIpForward             bool                // 是否启用 net.ipv4.ip_forward 功能
	DefaultIp                   net.IP              //	绑定容器端口时，默认使用的 Ip
	BridgeIface                 string              // 添加容器网络至已有的网桥的接口名
	BridgeIp                    string              // 创建默认网桥的 Ip 地址
	InterContainerCommunication bool                // 是否允许宿主机上的容器之间相互通信
	GraphDriver                 string              // daemon 使用的存储驱动
	GraphOptions                []string            // daemon 存储驱动配置选项
	ExecDriver                  string              // daemon 运行时使用的 exec 驱动
	Mtu                         int                 // 容器间通信网络的 mtu
	DisableNetwork              bool                // 是否启用 daemon 的网络模式
	EnableSelinuxSupport        bool                // 是否启用对 selinux 的支持
	context                     map[string][]string // 上下文
	// todo new in book
	Mirrors      []string // 指定的 dockergsh registry 镜像地址
	EnableIpMasq bool     // 是否启用 Ip 伪装技术
	FixedCIDR    string   // 指定默认网桥的子网地址
}

// InitialFlags todo init
// InitialFlags 将命令行选项添加到当前进程的顶级标志解析器。
func (config *Config) InitialFlags(context *cli.Context) {
	config.PidFile = context.String("pidfile")
	config.Root = context.String("graph")
	config.AutoRestart = context.Bool("restart")
	config.Dns = utils.Validates(context.StringSlice("dns"), utils.ValidateIPAddress)
	config.DnsSearch = utils.Validates(context.StringSlice("dns-search"), utils.ValidateDnsSearch)
	config.EnableIptables = context.Bool("iptables")
	config.EnableIpForward = context.Bool("ip-forward")
	config.DefaultIp = net.ParseIP(context.String("ip"))
	config.BridgeIp = context.String("bridge-ip")
	config.BridgeIface = context.String("bridge")
	config.InterContainerCommunication = context.Bool("inter-container-communication")
	config.GraphDriver = context.String("storage-driver")
	config.GraphOptions = context.StringSlice("storage-opts")
	config.ExecDriver = context.String("exec-driver")
	config.EnableSelinuxSupport = context.Bool("selinux-enabled")
	config.Mtu = context.Int("mtu")
}

// GetDefaultNetworkMtu 获取 default 路由的默认 mtu 配置
func GetDefaultNetworkMtu() int {
	// 获取到 default 路由网卡信息，即 ip route 的第一条 default 信息
	iface, err := networkdriver.GetDefaultRouteIface()
	if err != nil {
		return DefaultNetworkMtu
	}
	return iface.MTU
}
