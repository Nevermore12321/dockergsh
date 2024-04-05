package netlink

import (
	"io"
	"net"
	"syscall"
	"unsafe"
)

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
	// 1. 创建 socket
	socket, err := getNetlinkSocket()
	if err != nil {
		return nil, err
	}

	// 2. 创建 请求体
	// 创建 NlMsghdr 头
	// syscall.RTM_GETROUTE 表示请求方法，想要获取路由信息
	// syscall.NLM_F_DUMP 表示向内核发送的消息是一个请求获取整个数据集的消息，而不仅仅是单个条目。这通常用于获取整个路由表、接口表或其他数据集
	netlinkReq := newNetlinkRequest(syscall.RTM_GETROUTE, syscall.NLM_F_DUMP)

	// 添加消息数据体
	// syscall.AF_UNSPEC 没有指定地址族，且适合任何协议族的地址
	msg := newIfInfomsg(syscall.AF_UNSPEC)
	netlinkReq.AddData(msg)

	// 3. 发送消息给内核
	if err := socket.Send(netlinkReq); err != nil {
		return nil, err
	}

	// 4. 校验，并解析返回数据
	pid, err := socket.GetPid() // 获取端口 pid
	if err != nil {
		return nil, err
	}

	res := make([]Route, 0) // 初始化返回结果

outer: // 遍历所有返回消息
	for {
		returnData, err := socket.Receive() // 接收数据，[]NetlinkMessage
		if err != nil {
			return nil, err
		}

		for _, resMsg := range returnData { // 遍历所有的 NetlinkMessage，校验并转换成 Route
			if err := socket.CheckMessage(resMsg, netlinkReq.Seq, pid); err != nil {
				if err == io.EOF { // 如果已经读取完成，退出循环
					break outer
				}
				return nil, err // 如果出错，返回错误
			}

			// 如果返回的结果不是RTM_GETROUTE类型（获取路由信息的发送类型）
			if resMsg.Header.Type != syscall.RTM_GETROUTE {
				continue
			}

			var r Route

			// resMsg 是 NetlinkMessage 结构体，Data 是返回的数据信息，字节数组
			// 将其 [0:syscall.SizeofRtMsg] 字节数据，转成 RtMsg 类型
			rtMsg := (*RtMsg)(unsafe.Pointer(&resMsg.Data[0:syscall.SizeofRtMsg][0]))

			if rtMsg.Flags&syscall.RTM_F_CLONED != 0 {
				// 如果路由是通过另一个路由 clone 而来的，跳过
				continue
			}

			if rtMsg.Table != syscall.RT_TABLE_MAIN {
				// 如果不是主路由，跳过
				continue
			}

			if rtMsg.Family != syscall.AF_INET {
				// 如果不是 ipv4 路由，跳过
				continue
			}

			if rtMsg.Dst_len == 0 {
				// 如果目标地址长度为0,说明针对说有，即默认路由
				r.Default = true
			}

			// 将有效负载解析为 netlink 路由属性数组 []NetlinkRouteAttr
			attrs, err := syscall.ParseNetlinkRouteAttr(&resMsg)
			if err != nil {
				return nil, err
			}
			// 将 路由属性，转换成 IpNet 和 Interface
			for _, attr := range attrs {
				switch attr.Attr.Type { // NetlinkRouteAttr.Attr.Type 表示了路由属性的类型
				case syscall.RTA_DST: // RTA_DST 表示路由表项中的目标地址属性
					ip := attr.Value
					r.IPNet = &net.IPNet{
						IP:   ip,                                          // ip 是字节数组，例如 192.168.10.0，Ip 就表示为 [192,168,10,0]
						Mask: net.CIDRMask(int(rtMsg.Dst_len), len(ip)*8), // ones 参数表示子网掩码中连续的 1 的位数（即掩码位数），bits总长度
					}

				case syscall.RTA_OIF: // RTA_OIF 常量用于表示出口接口属性，即指定了数据包从哪个网络接口发送
					// attr.Value 的前四位表示了目的 IP 地址
					index := int(native.Uint32(attr.Value[0:4]))
					r.Iface, _ = net.InterfaceByIndex(index)
				}
			}

			if r.Default || r.IPNet != nil {
				res = append(res, r)
			}

		}
	}

	return res, nil
}
