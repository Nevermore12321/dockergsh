package network

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
	"net"
	"os/exec"
	"strings"
	"time"
)

type BridgeNetworkDriver struct {
}

func (bd *BridgeNetworkDriver) Name() string {
	return "bridge"
}

// Create 通过 linux-bridge 创建网络
// subnet: 子网网段、网关
// name: 子网网络名称
func (bd *BridgeNetworkDriver) Create(subnet string, name string) (*Network, error) {
	// 通过 net 包中的 net.ParaseCIDR 解析 subnet, 得到 网关和 Ip地址段
	// ip 为网关 ip
	ip, ipRange, err := net.ParseCIDR(subnet)
	if err != nil {
		log.Errorf("Linux Bridge create Network ParseCIDR error: %s", err)
		return nil, err
	}

	ipRange.IP = ip

	// 初始化网络
	network := &Network{
		Name:    name,
		IpRange: ipRange,
		Driver:  bd.Name(),
	}

	// 配置 Linux-Bridge
	err = bd.initBridge(network)
	if err != nil {
		log.Errorf("Linux Bridge create Network initBridge error: %s", err)
		return nil, err
	}

	return network, nil
}

// linux-bridge 设备初始化
func (bd *BridgeNetworkDriver) initBridge(network *Network) error {
	// 1. 创建 linux-bridge 虚拟设备
	bridgeName := network.Name
	err := createBridgeInterface(bridgeName)
	if err != nil {
		return fmt.Errorf("error add bridge： %s, Error: %v", bridgeName, err)
	}

	// 2. 设置 Bridge 设备的地址和路由
	gatewayIP := *network.IpRange
	gatewayIP.IP = network.IpRange.IP

	if err = setInterfaceIP(bridgeName, gatewayIP.String()); err != nil {
		return fmt.Errorf("error assigning address: %s on bridge: %s with an error of: %v", gatewayIP, bridgeName, err)
	}

	// 3. 启动 Bridge 设备
	if err = setInterfaceUp(bridgeName); err != nil {
		return fmt.Errorf("error set bridge up: %s, Error: %v", bridgeName, err)
	}

	// 4. 设置 Iptables 的 SNAT 规则
	if err = setupIptablesSNAT(bridgeName, network.IpRange); err != nil {
		return fmt.Errorf("error setting iptables for %s: %v", bridgeName, err)
	}

	return nil

}

// Linux-bridge 删除网络
func (bd *BridgeNetworkDriver) Delete(network *Network) error {
	return nil
}

// Linux-bridge 连接网络端点到新建的网络
func (bd *BridgeNetworkDriver) Connect(network *Network, endpoint *Endpoint) error {
	return nil
}

// Linux-bridge 将新建的网络端点删除，断开连接
func (bd *BridgeNetworkDriver) Disconnect(network *Network, endpoint *Endpoint) error {
	return nil
}

// 创建 linux-bridge 设备
func createBridgeInterface(bridgeName string) error {
	// 先检查是否已经存在了同名的 bridge 设备
	_, err := net.InterfaceByName(bridgeName)
	// 如果存在, 直接返回 nil, 如果报错不存在，返回错误
	if err == nil || !strings.Contains(err.Error(), "no such network interface") {
		return err
	}

	// 初始化一个 netlink 的 Link 对象，Link 的名字就是 Bridge 虚拟设备的名字
	// 首先创建 默认 Link 对象所需的选项对象 LinkAttrs，采用默认配置
	linkAttrs := netlink.NewLinkAttrs()
	linkAttrs.Name = bridgeName

	// 使用刚刚创建的 LinkAttrs 配置对象，创建 netlink 的 Bridge 对象
	bridgeDevice := &netlink.Bridge{LinkAttrs: linkAttrs}

	// 调用 netlink 的 LinkAdd 方法，创建 Bridge 虚拟网络
	// LinkAdd 方法就是用创建虚拟设备，其实就是调用 ip link add xxx
	if err := netlink.LinkAdd(bridgeDevice); err != nil {
		return fmt.Errorf("bridge creation failed for bridge %s: %v", bridgeName, err)
	}
	return nil
}

// 设置一个网络接口的 Ip 地址
// ifaceName: 网络名称
// rawIP: IP地址
func setInterfaceIP(interfaceName string, rawIP string) error {
	retries := 2
	var iface netlink.Link
	var err error

	// 重试三次
	for i := 0; i < retries; i++ {
		// 通过 netlink.LinkByName. 得到对应网络接口
		iface, err = netlink.LinkByName(interfaceName)
		if err == nil {
			break
		}
		log.Debugf("error retrieving new bridge netlink link [ %s ]... retrying", interfaceName)
		time.Sleep(2 * time.Second)
	}

	// 重试三次后，仍然失败，返回错误
	if err != nil {
		return fmt.Errorf("abandoning retrieving the new bridge link from netlink, Run [ ip link ] to troubleshoot the error: %v", err)
	}
	/*
		netlink.ParseIPNet 解析 IP 地址，是对 net.ParseCIDR 的封装，将返回值 IP 和 Net 整合
		netlink.ParseIPNet 返回值 IPNet 既包含了网段信息 192.168.0.0/24,也包含了原始的 IP 地址：192.168.0.1
	*/
	ipNet, err := netlink.ParseIPNet(rawIP)
	if err != nil {
		return err
	}

	/*
		通过 netlink.AddrAdd 方法，设置 IP 地址，相当于 ip addr add xxx dev xxx
		如果同时配置了地址所在网段信息，例如 192.168.0.0/24
		还会配置路由表 192.168.0.0/24 转发到这个 bridge 上
	*/
	addr := &netlink.Addr{
		IPNet:     ipNet,
		Peer:      ipNet,
		Label:     "",
		Flags:     0,
		Scope:     0,
		Broadcast: nil,
	}
	return netlink.AddrAdd(iface, addr)
}

// 设置网络接口为 UP 状态
func setInterfaceUp(interfaceName string) error {
	// 先查找 网络接口
	iface, err := netlink.LinkByName(interfaceName)
	if err != nil {
		return fmt.Errorf("error retrieving a link named [ %s ]: %v", interfaceName, err)
	}

	// 通过 netlink 的 LinkSetUp 方法，设置网络接口状态 UP，先当与 ip addr set xxx up
	err = netlink.LinkSetUp(iface)
	if err != nil {
		return fmt.Errorf("error enabling interface for %s: %v", interfaceName, err)
	}
	return nil
}

// 设置 iptables SNAT 规则，设置 iptables 对应 bridge 的 MASQUERADE 规则
// masquerade 规则，就是从服务器的网卡上自动获取网卡地址，来所 SNAT
func setupIptablesSNAT(bridgeName string, subnet *net.IPNet) error {
	// Go 没有直接操作 iptables 规则
	// 只能直接 linux 命令，iptables -t nat -A POSTROUTING -s <BridgeName> ! -o <BridgeName> -j MASQUERADE
	iptablesCmd := fmt.Sprintf("-t nat -A POSTROUTING -s %s ! -o %s -j MASQUERADE", subnet.String(), bridgeName)
	cmd := exec.Command("iptables", strings.Split(iptablesCmd, " ")...)
	// 执行 iptables 规则
	output, err := cmd.Output()
	if err != nil {
		log.Errorf("iptables Output, %v", output)
	}
	return nil
}
