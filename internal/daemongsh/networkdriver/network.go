package networkdriver

import (
	"errors"
	"fmt"
	"github.com/Nevermore12321/dockergsh/external/libcontainer/netlink"
	"net"
)

// network 默认配置
var (
	DefaultNetworkMtu    = 1500 // mtu 默认值
	DisableNetworkBridge = "none"
)

// error 定义
var (
	ErrNoDefaultRoute = errors.New("no default route") // 没有找到 default 路由信息
)

var (
	// todo  libcontainer 实现 NetworkGetRoutes 方法
	networkGetRoutesFct = netlink.NetworkGetRoutes
)

// 获取默认的路由，类似 ip route 的第一条 default路由
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
