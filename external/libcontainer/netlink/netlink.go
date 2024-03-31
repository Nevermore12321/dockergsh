package netlink

import (
	"errors"
	"net"
)

var (
	ErrShortResponse = errors.New("got short response from netlink") // netlink 返回结果长度小于 NLMSG 头长度，即错误
	ErrWrongSockType = errors.New("wrong socket type")               // socket 中的 pid 解析类型错误
)

/*
Route 路由信息，表示一条路由信息
*/
type Route struct {
	*net.IPNet
	Iface   *net.Interface // Iface 表示网络接口名称和索引之间的映射。它还表示网络接口设施信息。
	Default bool           // 是否是默认路由
}

// IfAddr 网卡信息
type IfAddr struct {
	Iface *net.Interface
	IP    net.IP
	IPNet *net.IPNet
}
