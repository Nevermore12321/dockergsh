package netlink

import "syscall"

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
	// todo
	return nil
}
