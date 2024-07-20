package netlink

import (
	"encoding/binary"
	"sync/atomic"
	"syscall"
	"unsafe"
)

// 全局的请求序列号 seq
var nextSeqNr uint32

// Binary 包实现了简单的数字和字节序列之间的转换以及变量的编码与解码。
// ByteOrder 接口将字节片转换为 16 位、32 位或 64 位无符号整数。
// native 表示一个字节序接口
var native binary.ByteOrder

// 初始化函数，用于判断当前机器是大端字节序还是小端字节序
func init() {
	// 初始化一个十六进制数
	var x uint32 = 0x01020304

	// 先将 uint32 指针转化成 byte 类型指针，byte 只有8位，也就是取的最低的8位
	if *(*byte)(unsafe.Pointer(&x)) == 0x01 {
		// 在大端字节序中，高位字节存储在低地址，低位字节存储在高地址
		native = binary.BigEndian
	} else {
		// 在小端字节序中，低位字节存储在低地址，高位字节存储在高地址
		native = binary.LittleEndian
	}
}

// NetlinkRequestData netlink 报文的消息数据 payload 接口，必须实现下面两个方法
type NetlinkRequestData interface {
	Len() int             // 返回消息体长度
	ToWireFormat() []byte // 编码成二进制字节数组
}

// IfInfomsg 重新创建是为了添加方法
// IfInfomsg 表示网络接口的信息，包括网络接口的索引、名称、类型等属性
// 对应于 Linux 内核中的 struct ifinfomsg 结构体，用于获取和设置网络接口的属性
type IfInfomsg struct {
	syscall.IfInfomsg
}

// IfInfomsg 实例化
// 地址族（family）：地址族指定了套接字所使用的地址类型
func newIfInfomsg(family uint8) *IfInfomsg {
	return &IfInfomsg{
		IfInfomsg: syscall.IfInfomsg{
			Family: family,
		},
	}
}

// ToWireFormat 将 IfInfomsg 报文转成二进制字节数组，实现 NetlinkRequestData 接口
func (ifmsg *IfInfomsg) ToWireFormat() []byte {
	length := syscall.SizeofIfInfomsg // IfInfomsg 消息长度
	dataBytes := make([]byte, length) // 根据 IfInfomsg 消息长度初始化字节数组

	// 因为 IfInfomsg 消息的前两个元素(Family，X__ifi_pad - 用于填充结构体以保持与内核数据结构的对齐)是 uint8 类型，也就是8位，分别占一个字节，因此直接设置
	dataBytes[0] = ifmsg.Family
	dataBytes[1] = 0
	native.PutUint16(dataBytes[2:4], ifmsg.Type)          // [2,3] 位表示 IfInfomsg.Type
	native.PutUint32(dataBytes[4:8], uint32(ifmsg.Index)) // [4,7] 位表示 IfInfomsg.Index
	native.PutUint32(dataBytes[8:12], ifmsg.Flags)        // [8,11] 位表示 IfInfomsg.Flags
	native.PutUint32(dataBytes[12:16], ifmsg.Change)      // [12,15] 位表示 IfInfomsg.Change

	return dataBytes
}

// Len 返回 IfInfomsg 长度，实现 NetlinkRequestData 接口
func (ifmsg *IfInfomsg) Len() int {
	return syscall.SizeofIfInfomsg
}

// RtMsg 是 rtnetlink 消息的一个结构体
// 用于表示从内核接收或发送的网络路由信息。该结构体包含了与路由相关的各种属性，如目的地址、网关地址、路由表索引等
//   - Family: 表示地址族，如 IPv4 或 IPv6。
//   - Dst_len: 目的地址长度。
//   - Src_len: 源地址长度。
//   - Tos: 服务类型。
//   - Table: 路由表索引。
//   - Protocol: 使用的协议。
//     ...
type RtMsg struct {
	syscall.RtMsg
}

// Table:RT_TABLE_MAIN
//   - 主路由表，使用传统命令route -n所看到的路由表就是main的内容
//   - linux系统在默认情况下使用这份路由表的内容来传输数据包。正常情况下，只有配置好网卡的网络设置，便会自动生成main路由表的内容
//
// Protocol:RTPROT_BOOT
//   - 网卡启动时建立的路由
//
// Scope:RT_SCOPE_UNIVERSE
//   - 全局路由
//
// Type:RTN_UNICAST
//   - 单播（unicast）路由，即指向单个目的地址的路由
func newRtMsg() *RtMsg {
	return &RtMsg{
		RtMsg: syscall.RtMsg{
			Table:    syscall.RT_TABLE_MAIN,
			Protocol: syscall.RTPROT_BOOT,
			Scope:    syscall.RT_SCOPE_UNIVERSE,
			Type:     syscall.RTN_UNICAST,
		},
	}
}

// ToWireFormat 将 RtMsg 报文转成二进制字节数组，实现 NetlinkRequestData 接口
func (rtMsg *RtMsg) ToWireFormat() []byte {
	// ToWireFormat 的逻辑都一致
	length := syscall.SizeofRtMsg
	dataBytes := make([]byte, length)
	dataBytes[0] = rtMsg.Family
	dataBytes[1] = rtMsg.Dst_len
	dataBytes[2] = rtMsg.Src_len
	dataBytes[3] = rtMsg.Tos
	dataBytes[4] = rtMsg.Table
	dataBytes[5] = rtMsg.Protocol
	dataBytes[6] = rtMsg.Scope
	dataBytes[7] = rtMsg.Type
	native.PutUint32(dataBytes[8:12], rtMsg.Flags) // Flags 标志是 int32,有 4B，其他标志都是 uint8,只有1字节
	return dataBytes
}

// Len 返回 RtMsg 长度，实现 NetlinkRequestData 接口
func (ifmsg *RtMsg) Len() int {
	return syscall.SizeofRtMsg
}

// 上面是各类 request data ，以及实现方法
// ===============================================
// 下面是完整的 request 信息

// NetlinkRequest netlink 协议发送的消息，Netlink的报文由消息头和消息体构成
// NlMsghdr 结构体的定义通常与 C 语言中的 struct nlmsghdr 类似，它包含了以下字段：
//   - Len：表示整个消息的长度，单位是字节。包括了Netlink消息头本身
//   - Type：表示消息的类型，即是数据还是控制消息。例如：
//     1. NLMSG_NOOP - 空消息，什么也不做；
//     2. NLMSG_ERROR - 指明该消息中包含一个错误；
//     3. NLMSG_DONE - 如果内核通过Netlink队列返回了多个消息，那么队列的最后一条消息的类型为NLMSG_DONE，其余所有消息的nlmsg_flags属性都被设置NLM_F_MULTI位有效。
//     4. NLMSG_OVERRUN - 暂时没用到。
//   - Flags：表示消息的附加标志位，用于指示消息的一些附加信息。如上面提到的 NLM_F_MULTI
//   - Seq：表示消息的序列号，用于区分不同的消息。
//   - Pid：表示消息的发送者Socket的Port号。
type NetlinkRequest struct {
	syscall.NlMsghdr                      // 消息头
	Data             []NetlinkRequestData // 数据体
}

// NetlinkRequest 实例化
func newNetlinkRequest(proto, flags int) *NetlinkRequest {
	return &NetlinkRequest{
		NlMsghdr: syscall.NlMsghdr{
			Len:   uint32(syscall.NLA_HDRLEN), // 此时初始化的是 header 的长度
			Type:  uint16(proto),
			Flags: syscall.NLM_F_REQUEST | uint16(flags),
			Seq:   atomic.AddUint32(&nextSeqNr, 1), // 每次请求都有一个 Seq 序号
		},
	}
}

// ToWireFormat 将 netlink 发送的完整报文转成二进制字节数组
func (req *NetlinkRequest) ToWireFormat() []byte {
	// 这里计算的长度，应包括 header + data
	length := req.Len
	dataBytes := make([][]byte, len(req.Data)) // 二维byte数组缓存区，一共 len() 行

	for index, data := range req.Data {
		// 每一个Data转成字节数组，存入缓存区
		dataBytes[index] = data.ToWireFormat()
		// 将每一个 DATA 算入总长度
		length += uint32(len(dataBytes[index]))
	}

	// b 是最终生成的字节消息体
	b := make([]byte, length)

	// 构建消息体，按照特定的字节序，将 header NlMsghdr 写入到字节数组b中
	native.PutUint32(b[0:4], length)    // [0,3] 位表示 NlMsghdr.Len
	native.PutUint16(b[4:6], req.Type)  // [4,5] 位表示 NlMsghdr.Type
	native.PutUint16(b[6:8], req.Flags) // [6,7] 位表示 NlMsghdr.Flags
	native.PutUint32(b[8:12], req.Seq)  // [8,12] 位表示 NlMsghdr.Seq
	native.PutUint32(b[12:16], req.Pid) // [12,16] 位表示 NlMsghdr.Pid

	// 下一个字节位置到 16
	nextIndex := 16

	// 按照特定的字节序，继续将 Data 写入到字节数组b中
	for _, data := range dataBytes {
		// 因为 data 已经转成字节数组了，直接拷贝
		copy(b[nextIndex:], data)
		nextIndex += len(data)
	}

	return b
}

// AddData 追加消息
func (req *NetlinkRequest) AddData(data NetlinkRequestData) {
	if data != nil {
		req.Data = append(req.Data, data)
	}
}

// 设置mac地址的请求结构
type ifreqHwAddr struct {
	ifrnName   [IFNAMESIZE]byte // 网卡名称
	ifruHwAddr syscall.RawSockaddr
}

// IfAddrMsg 是 rtnetlink 消息的一个结构体
// 用于与操作系统进行低级别的网络接口操作,syscall.IfAddrmsg包括：
// - Family：地址族（address family），例如 AF_INET 表示 IPv4，AF_INET6 表示 IPv6。
// - Prefixlen：地址的前缀长度（prefix length），即子网掩码的长度。
// - Flags：与地址相关的标志（flags），例如是否是永久地址等。
// - Scope：地址的作用域（scope），表示地址的可达性范围，例如本地、链路、全局等。
// - Index：网络接口的索引（index），用于标识具体的网络接口。
type IfAddrMsg struct {
	syscall.IfAddrmsg
}

// 实例化 IfAddrMsg 结构，选择地址族
func newIfAddrMsg(family int) *IfAddrMsg {
	return &IfAddrMsg{
		IfAddrmsg: syscall.IfAddrmsg{
			Family: uint8(family),
		},
	}
}

// ToWireFormat 将 IfAddrMsg 报文转成二进制字节数组，实现 NetlinkRequestData 接口
func (msg *IfAddrMsg) ToWireFormat() []byte {
	length := syscall.SizeofIfAddrmsg
	dataBytes := make([]byte, length)
	dataBytes[0] = msg.Family
	dataBytes[1] = msg.Prefixlen
	dataBytes[2] = msg.Flags
	dataBytes[3] = msg.Scope
	native.PutUint32(dataBytes[4:8], msg.Index)
	return dataBytes
}

// Len 返回 IfAddrMsg 长度，实现 NetlinkRequestData 接口
func (msg *IfAddrMsg) Len() int {
	return syscall.SizeofIfAddrmsg
}

// RtAttr 在配置网络接口地址时，需要创建多个 RtAttr 属性，并将其添加到 Netlink 消息中发送给内核。内核解析 Netlink 消息并根据属性进行相应的配置操作。
type RtAttr struct {
	syscall.RtAttr                      // 包括 Type(属性的类型) 和 Len（属性长度）
	Data           []byte               // 属性的数据
	children       []NetlinkRequestData // 子请求
}

// 实例化 Rtattr
func newRtAttr(attrType int, data []byte) *RtAttr {
	return &RtAttr{
		RtAttr: syscall.RtAttr{
			Type: uint16(attrType),
		},
		Data:     data,
		children: []NetlinkRequestData{},
	}
}

// 按照 RTA_ALIGNTO 字节对齐
func rtaAlignOf(attrLen int) int {
	// - RTA_ALIGNTO 是一个常量，用于指定属性对齐的边界值。它通常定义为 4 字节
	// - 将长度 len 加上 RTA_ALIGNTO - 1，确保总长度大于或等于 RTA_ALIGNTO 的倍数。
	// - 使用按位与操作 & 和取反操作 ^(syscall.RTA_ALIGNTO - 1)，将总长度调整为 RTA_ALIGNTO 的倍数。
	return (attrLen + syscall.RTA_ALIGNTO - 1) & ^(syscall.RTA_ALIGNTO - 1)
}

// Len 返回 RtAttr 长度，实现 NetlinkRequestData 接口
func (a *RtAttr) Len() int {
	if len(a.children) == 0 {
		return (syscall.SizeofRtAttr + len(a.Data))
	}
	length := 0
	// 遍历所有子请求，计算长度
	for _, child := range a.children {
		length += child.Len()
	}

	// 再加上 syscall.RtAttr 长度 和 Data 的长度
	length += syscall.SizeofRtAttr
	length += len(a.Data)
	// 按照 对齐
	return rtaAlignOf(length)
}

// ToWireFormat 将 IfAddrMsg 报文转成二进制字节数组，实现 NetlinkRequestData 接口
func (a *RtAttr) ToWireFormat() []byte {
	length := a.Len()
	buf := make([]byte, rtaAlignOf(length))

	if a.Data != nil { // 如果有数据
		copy(buf[4:], a.Data) // 从第四位开始都是属性数据
	} else { // 如果没有数据，计算子请求数据
		// 子请求的二进制数组
		next := 4
		for _, child := range a.children {
			childBuf := child.ToWireFormat()
			copy(buf[next:], childBuf)
			next += rtaAlignOf(len(childBuf))
		}
	}

	// 长度写入 0,1 位置
	if l := uint16(length); l > 0 {
		native.PutUint16(buf[0:2], l)
	}
	native.PutUint16(buf[2:4], a.Type)
	return buf
}

// IfInfoMsg RTM_NEWLINK 创建/删除/修改 网卡设备的消息结构
// - Family (uint8): 地址族，通常是 AF_INET (IPv4) 或 AF_INET6 (IPv6)，也可以是其他值来表示不同类型的网络。
// - Type (uint16): 接口类型，表示网络接口的类型（例如，ARPHRD_ETHER 表示以太网接口）。
// - Index (int32): 接口索引，这是系统分配的唯一标识符，用于标识具体的网络接口。
// - Flags (uint32): 接口标志，表示接口的状态或特性（例如，IFF_UP 表示接口已启动，IFF_RUNNING 表示接口正在运行）。
// - Change (uint32): 改变掩码，用于指示哪些标志需要更改。
type IfInfoMsg struct {
	syscall.IfInfomsg
}

// 实例化 IfInfomsg 结构体M
func newIfInfoMsg(family int) *IfInfomsg {
	return &IfInfomsg{
		IfInfomsg: syscall.IfInfomsg{
			Family: uint8(family),
		},
	}
}

func (msg *IfInfoMsg) Len() int {
	return syscall.SizeofIfInfomsg
}

func (msg *IfInfoMsg) ToWireFormat() []byte {
	length := syscall.SizeofIfInfomsg
	dataBytes := make([]byte, length)
	dataBytes[0] = msg.Family
	dataBytes[1] = 0 // 填充字节 X__ifi_pad 字段
	native.PutUint16(dataBytes[2:4], msg.Type)
	native.PutUint32(dataBytes[4:8], uint32(msg.Index))
	native.PutUint32(dataBytes[8:12], msg.Flags)
	native.PutUint32(dataBytes[12:16], msg.Change)
	return dataBytes
}
