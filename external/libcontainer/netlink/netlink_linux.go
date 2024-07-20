package netlink

import (
	"fmt"
	"io"
	"math/rand"
	"net"
	"syscall"
	"time"
	"unsafe"
)

const (
	IFNAMESIZE   = 16
	SIOC_BRADDBR = 0x89a0 // ioctl 添加 bridge 的请求码
)

var rnd = rand.New(rand.NewSource(time.Now().UnixNano())) // 随机数种子

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

	defer socket.Close()

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
			if resMsg.Header.Type != syscall.RTM_NEWROUTE {
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

// CreateBridge 创建网桥
func CreateBridge(name string, setMacAddr bool) error {
	if len(name) >= IFNAMESIZE { // 如果名称长度超出 16 个字符，报错
		return fmt.Errorf("interface name %s too long", name)
	}

	// 获取网络的套接字，
	s, err := getIfSocket()
	if err != nil {
		return err
	}
	defer syscall.Close(s)

	// 将 name 字符串转成字节数组指针
	nameBytePtr, err := syscall.BytePtrFromString(name)
	if err != nil {
		return err
	}

	// 通过调用 syscall.SYS_IOCTL 中断，来调用 系统的 ioctl 函数
	// ioctl 需要三个参数，int ioctl(int d, int request, ...);
	//	- 第一个参数d指定一个由open/socket创建的文件描述符，即socket套接字
	//	- 第二个参数request指定操作的类型，即对该文件描述符执行何种操作，设备相关的请求的代码
	//	- 第三个参数为一块内存区域，通常依赖于request指定的操作类型
	if _, _, err = syscall.Syscall(syscall.SYS_IOCTL, uintptr(s), SIOC_BRADDBR, uintptr(unsafe.Pointer(nameBytePtr))); err != nil {
		return err
	}

	if setMacAddr {
		return setMacAddress(name, randMacAddress())
	}

	return nil
}

// 随机生成 mac 地址
func randMacAddress() string {
	hw := make(net.HardwareAddr, 6)
	for i := 0; i < 6; i++ {
		hw[i] = byte(rnd.Intn(255))
	}

	hw[0] &^= 0x1 // 将 mac 的第一 byte 的最低位置 0, clear multicast bit
	hw[0] |= 0x2  // 将 mac 的第一 byte 的第二位置 1, set local assignment bit (IEEE802)
	return hw.String()
}

// 判断ip是 ipv4 地址还是 ipv6 地址
func getIpFamily(ip net.IP) int {
	if len(ip) <= net.IPv4len {
		return syscall.AF_INET
	}
	if ip.To4() != nil {
		return syscall.AF_INET
	}
	return syscall.AF_INET6
}

// 给指定网卡设置 mac 地址
func setMacAddress(name, add string) error {
	if len(name) >= IFNAMESIZE { // 如果名称长度超出 16 个字符，报错
		return fmt.Errorf("interface name %s too long", name)
	}

	// 解析 mac 地址
	hw, err := net.ParseMAC(add)
	if err != nil {
		return err
	}

	// 获取网络套接字
	s, err := getIfSocket()
	if err != nil {
		return err
	}
	defer syscall.Close(s)

	ifr := ifreqHwAddr{}
	// 以太网协议，用于操作以太网设备
	ifr.ifruHwAddr.Family = syscall.ARPHRD_ETHER
	// 要操作的以太网设备名称
	copy(ifr.ifrnName[:len(ifr.ifrnName)-1], name)

	// 设置的 MAC 地址
	for i := 0; i < 6; i++ {
		ifr.ifruHwAddr.Data[i] = ifrDataByte(hw[i])
	}

	// syscall.SIOCSIFHWADDR ioctl 函数的设置mac地址功能
	_, _, err = syscall.Syscall(syscall.SYS_IOCTL, uintptr(s), syscall.SIOCSIFHWADDR, uintptr(unsafe.Pointer(&ifr)))
	if err != nil {
		return err
	}
	return nil
}

// 网卡IP配置的相关操作action公共抽象(通过 netlink 机制)
func networkLinkIpAction(action, flag int, ifa IfAddr) error {
	// 1. create netlink socket
	socket, err := getNetlinkSocket()
	if err != nil {
		return err
	}
	defer socket.Close()

	// 2. ipv4 or ipv6?
	family := getIpFamily(ifa.IP)

	// 3. create netlink request
	wb := newNetlinkRequest(action, flag)

	// 4. create send message
	msg := newIfAddrMsg(family)
	msg.Index = uint32(ifa.Iface.Index) // 网络接口的索引
	prefixLen, _ := ifa.IPNet.Mask.Size()
	msg.Prefixlen = uint8(prefixLen)

	// 5. add msg to request
	wb.AddData(msg)

	var ipData []byte
	if family == syscall.AF_INET { // ip 地址
		ipData = ifa.IP.To4()
	} else {
		ipData = ifa.IP.To16()
	}
	// IFA_ADDRESS表示的是对端的地址，IFA_LOCAL表示的是本地ip地址
	// 在配置了支持广播的接口上，与IFA_LOCAL一样，同样表示本地ip地址
	localData := newRtAttr(syscall.IFA_LOCAL, ipData)
	wb.AddData(localData)

	addrData := newRtAttr(syscall.IFA_ADDRESS, ipData)
	wb.AddData(addrData)

	// 6. send msg
	if err := socket.Send(wb); err != nil {
		return err
	}

	return socket.HandleAck(wb.Seq)
}

// NetworkLinkAddIp 将 IP 地址添加到接口设备。即命令： ip addr add $ip/$ipNet dev $iface
func NetworkLinkAddIp(iface *net.Interface, ip net.IP, ipNet *net.IPNet) error {
	// action 是 RTM_NEWADDR 表示添加addr地址信息
	// flags 如下：
	// - syscall.NLM_F_CREATE: 指示要创建一个新的对象（如果对象不存在）。
	// - syscall.NLM_F_EXCL: 与 NLM_F_CREATE 一起使用，指示如果对象已经存在，则请求失败。
	// - syscall.NLM_F_ACK: 指示请求应接收确认（ACK）响应。
	return networkLinkIpAction(
		syscall.RTM_NEWADDR,
		syscall.NLM_F_CREATE|syscall.NLM_F_EXCL|syscall.NLM_F_ACK,
		IfAddr{iface, ip, ipNet},
	)
}

// NetworkLinkUp 启动特定的网络接口。即命令： ip link set dev name up
func NetworkLinkUp(iface *net.Interface) error {
	// 1. create netlink socket
	socket, err := getNetlinkSocket()
	if err != nil {
		return err
	}
	defer socket.Close()

	// 2. create request: RTM_NEWLINK - 创建网卡； NLM_F_ACK -Reply with ack
	// RTM_NEWLINK 请求的消息结构放在 ifinfomsg 中
	wb := newNetlinkRequest(syscall.RTM_NEWLINK, syscall.NLM_F_ACK)

	// 3. create message
	// AF_UNSPEC 表示不指定特定的地址族，既可以是 IPv4，也可以是 IPv6
	msg := newIfInfoMsg(syscall.AF_UNSPEC)
	msg.Index = int32(iface.Index)
	msg.Flags = syscall.IFF_UP // 用于表示设置网络接口的“启动”状态标志
	msg.Change = syscall.IFF_UP

	// 4. add msg data to request
	wb.AddData(msg)

	// 5. send msg to socket
	if err := socket.Send(wb); err != nil {
		return err
	}

	// 5. check response and ack
	return socket.HandleAck(wb.Seq)
}
