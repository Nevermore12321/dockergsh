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

// NetlinkRequest 实例化
func newNetlinkRequest(proto, flags int) *NetlinkRequest {
	return &NetlinkRequest{
		NlMsghdr: syscall.NlMsghdr{
			Len:   uint32(syscall.NLA_HDRLEN), // 此时初始化的是 header 的长度
			Type:  uint16(proto),
			Flags: syscall.NLM_F_REQUEST | uint16(flags),
			Seq:   atomic.AddUint32(&nextSeqNr, 1),
		},
	}
}
