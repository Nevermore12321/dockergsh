package networkdriver

import (
	"fmt"
	"github.com/Nevermore12321/dockergsh/external/libcontainer/netlink"
	"net"
)

var (
	//  libcontainer 实现 NetworkGetRoutes 方法
	networkGetRoutesFct = netlink.NetworkGetRoutes
)

// GetDefaultRouteIface 获取默认的路由，类似 ip route 的第一条 default路由
func GetDefaultRouteIface() (*net.Interface, error) {
	// 通过 libcontainer 的 netlink 包，获取 ip route 获取所有的路由信息
	routes, err := networkGetRoutesFct()
	if err != nil {
		return nil, fmt.Errorf("unable to get routes: %v", err)
	}

	// 遍历路由，找到第一条 default 路由
	for _, route := range routes {
		if route.Default {
			return route.Iface, nil
		}
	}

	// 如果没有找到，返回 错误 ErrNoDefaultRoute
	return nil, ErrNoDefaultRoute
}

// GetIfaceAddr 通过网络接口名称返回网络接口的 IPv4 地址
func GetIfaceAddr(name string) (net.Addr, error) {
	iface, err := net.InterfaceByName(name) // 通过接口名称获取接口信息
	if err != nil {
		return nil, err
	}
	// Addrs 返回特定接口的单播接口地址列表
	addrs, err := iface.Addrs()
	if err != nil {
		return nil, err
	}

	// 获取所有 ipv4 地址
	var addrs4 []net.Addr
	for _, addr := range addrs {
		ip := (addr.(*net.IPNet)).IP
		if ipv4 := ip.To4(); len(ipv4) == net.IPv4len {
			addrs4 = append(addrs4, addr)
		}
	}

	// 判断是否有绑定 ipv4 地址
	switch {
	case len(addrs4) == 0: // 没有绑定 ip 地址，报错
		return nil, fmt.Errorf("interface %v has no IP addresses", name)
	case len(addrs4) > 1: // 如果绑定多个 ip 地址，使用第一个
		fmt.Printf("Interface %v has more than 1 IPv4 address. Defaulting to using %v\n", name, (addrs4[0].(*net.IPNet)).IP)
	}

	return addrs4[0], nil
}
