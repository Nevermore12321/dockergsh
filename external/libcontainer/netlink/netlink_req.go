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
