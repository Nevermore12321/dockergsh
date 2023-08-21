package daemon

import (
	"github.com/urfave/cli/v2"
	"net"
)

type Config struct {
	PidFile                     string              // daemongsh 所属进程的 PID 文件
	Root                        string              // dockergsh 运行时所使用的 root 路径
	AutoRestart                 bool                // 是否支持创建的容器自动重启
	Dns                         []string            // daemongsh 创建容器的默认 DNS server 地址
	DnsSearch                   []string            // dockergsh 使用指定的 DNS 查找地址
	EnableIptables              bool                // 是否启用 daemongsh 的 iptables 功能
	EnableIpForward             bool                // 是否启用 net.ipv4.ip_forward 功能
	DefaultIp                   net.IP              //	绑定容器端口时，默认使用的 Ip
	BridgeIface                 string              // 添加容器网络至已有的网桥的接口名
	BridgeIp                    string              // 创建默认网桥的 Ip 地址
	InterContainerCommunication bool                // 是否允许宿主机上的容器之间相互通信
	GraphDriver                 string              // daemongsh 使用的存储驱动
	GraphOptions                []string            // daemongsh 存储驱动配置选项
	ExecDriver                  string              // daemongsh 运行时使用的 exec 驱动
	Mtu                         int                 // 容器间通信网络的 mtu
	DisableNetwork              bool                // 是否启用 daemongsh 的网络模式
	EnableSelinuxSupport        bool                // 是否启用对 selinux 的支持
	context                     map[string][]string // 上下文
	// todo new in book
	Mirrors      []string // 指定的 dockergsh registry 镜像地址
	EnableIpMasq bool     // 是否启用 Ip 伪装技术
	FixedCIDR    string   // 指定默认网桥的子网地址
}

func (config *Config) InitialFlags(context *cli.Context) {
	config.PidFile = context.String("pidfile")
	config.Root = context.String("graph")
	config.AutoRestart = context.Bool("pidfile")
	config.Dns = context.StringSlice("pidfile")
	config.DnsSearch = context.StringSlice("pidfile")
	config.EnableIptables = context.Bool("pidfile")
	config.EnableIpForward = context.Bool("pidfile")
	//config.DefaultIp = context.StringSlice("pidfile")
	config.PidFile = context.String("pidfile")
	config.PidFile = context.String("pidfile")
	config.PidFile = context.String("pidfile")
	config.PidFile = context.String("pidfile")
	config.PidFile = context.String("pidfile")
	config.PidFile = context.String("pidfile")
	config.PidFile = context.String("pidfile")
	config.PidFile = context.String("pidfile")
}
