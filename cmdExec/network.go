package cmdExec

import (
	"github.com/Nevermore12321/dockergsh/network"
	log "github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
	"net"
)

// 表示一个网卡对应的网络信息
// 网络驱动是一个网络功能的中间件，不同的驱动对网络的创建、连接、销毁的策略不同，通过指定不同的 network driver来定义哪个驱动作为网络配置
type Network struct {
	Name    string     // 网络名称
	IpRange *net.IPNet //地址段
	Driver  string     // 网络驱动器名称
}

// 网络端点
// 网络端点是用于连接容器和网络的，保证容器内部与网络的通信
// 其实就相当于 veth pair 设备
type Endpoint struct {
	Id          string           `json:"id"`          // 网络端点 Id
	Device      netlink.Veth     `json:"device"`      // veth 设备
	IPAddress   net.IP           `json:"ip"`          // ip 地址
	MacAddress  net.HardwareAddr `json:"mac"`         // mac 地址
	PortMapping []string         `json:"portMapping"` // 端口映射
	Network     *Network         `json:"network"`     // 网络信息
}

// 创建网络
// driver: 网络驱动 driver，负责网络的创建等动作
// subnet: 创建网络的子网信息，也就是 网段
// name: 创建网络的名称
func CreateNetwork(driver, subnet, name string) (err error) {
	// net 标准库中的 ParseCIDR 函数将 CIDR 转成 net.IPNet 对象
	// IPNet 其实就是两个字段，一个网段IP，一个是掩码，例如 ParseCIDR("192.0.2.1/24") 返回IP地址 192.0.2.1 和掩码 192.0.2.0/24
	_, cidr, err := net.ParseCIDR(subnet)
	if err != nil {
		log.Errorf("Subnet convert to IPNet err: %v", err)
		return err
	}

	// 创建一个网络，那么必须有网关，需要在该网段中申请一个 ip 作为网关地址
	// 使用 ipam 包，获取网段中第一个 ip 作为网关地址，例如 192.168.1.1
	gatewayIp, err := network.IpAllocator.Allocate(cidr)
	if err != nil {
		return
	}
	cidr.IP = gatewayIp
	return

}
