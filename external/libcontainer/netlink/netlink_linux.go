package netlink

// NetworkGetRoutes 返回 ipv4 上所有当前路由子网的 IPNet 数组
/* 	这与“ip route”输出的第一列类似
```
❯ ip route                                                                                                                                 ─╯
default via 192.168.0.1 dev wlp4s0 proto dhcp src 192.168.0.100 metric 600
169.254.0.0/16 dev virbr0 scope link metric 1000 linkdown
172.17.0.0/16 dev docker0 proto kernel scope link src 172.17.0.1 linkdown
192.168.0.0/24 dev wlp4s0 proto kernel scope link src 192.168.0.100 metric 600
192.168.122.0/24 dev virbr0 proto kernel scope link src 192.168.122.1 linkdown
```
*/
func NetworkGetRoutes() ([]Route, error) {
	// 创建 socket
	socket, err := getNetlinkSocket()
}
