package netlink

import (
	"fmt"
	"io"
	"syscall"
)

type NetlinkSocket struct {
	fd int                     // 套接字描述符
	sa syscall.SockaddrNetlink // Sockaddr
}

// getNetlinkSocket 创建 NetlinkSocket 实例对象
// NetlinkSocket.sa 中定义了 Sockaddr，目标地址是 pid=0
func getNetlinkSocket() (*NetlinkSocket, error) {
	// 1. 创建 socket，使用 int socket(int domain, int type, int protocol); 系统调用
	//	domain：表示套接字的协议域（或地址族），指定了套接字通信所使用的协议类型。常见的协议族包括：
	//    - AF_INET：IPv4 协议族
	//    - AF_INET6：IPv6 协议族
	//    - AF_UNIX：UNIX 域（本地域）套接字
	//    - AF_PACKET：用于原始网络数据包通信
	// 	  - AF_NETLINK：netlink 协议族
	//
	//	type：表示套接字的类型，指定了套接字的通信语义和数据传输方式。常见的套接字类型包括：
	//    - SOCK_STREAM：流套接字（面向连接），提供可靠的、基于字节流的、全双工的数据传输。
	//    - SOCK_DGRAM：数据报套接字（无连接），提供不可靠的、无连接的、固定最大长度的数据传输。
	//    - SOCK_RAW：原始套接字，用于直接访问底层网络协议，通常需要特权权限。
	//
	//	protocol：表示套接字使用的具体协议。在大多数情况下，可以将此参数设置为 0，表示使用默认协议。
	//		例如，对于 AF_INET 和 SOCK_STREAM 类型的套接字，通常使用 TCP 协议，对于 AF_INET 和 SOCK_DGRAM 类型的套接字，通常使用 UDP 协议
	//

	// 1. 创建一个 socket，协议族是 AF_NETLINK，并且使用原始数据套接字
	// NETLINK_ROUTE 用户空间的路由守护程序之间的通讯通道，比如BGP,OSPF,RIP以及内核数据转发模块。用户态的路由守护程序通过此类型的协议来更新内核中的路由表。
	fd, err := syscall.Socket(syscall.AF_NETLINK, syscall.SOCK_RAW, syscall.NETLINK_ROUTE)
	if err != nil {
		return nil, err
	}

	// 2. 初始化 socket 描述符
	s := &NetlinkSocket{
		fd: fd,
	}
	// 3. 初始化 sa，SockaddrNetlink(Sockaddr)，即:
	// 		- pid: Socket的Port号，即要发送给哪个Socket，就指定为它的Port号，默认初始化 0（0 表示内核 socket）
	// 		- Groups: 是要发送的多播组，即将Netlink消息发送给哪些多播组，默认初始化 0 (表示 单播)
	//		- Family: 协议族，必须指定为 AF_NETLINK
	s.sa.Family = syscall.AF_NETLINK

	// 4. 将 socket 描述符与 SockaddrNetlink 绑定
	if err := syscall.Bind(fd, &s.sa); err != nil { // bind failed
		syscall.Close(fd)
		return nil, err
	}

	return s, nil
}

// Close 关闭套接字
func (s *NetlinkSocket) Close() {
	syscall.Close(s.fd)
}

// Send 发送消息给当前 socket(内核socket)，目标是 NetlinkSocket.sa
// 应用层向内核传递消息可以使用 sendto() 或 sendmsg() 函数
// 其中 sendmsg 函数需要应用程序手动封装 msghdr 消息结构，而 sendto() 函数则会由内核代为分配
func (s *NetlinkSocket) Send(request *NetlinkRequest) error {
	return syscall.Sendto(s.fd, request.ToWireFormat(), 0, &s.sa)
}

// Receive 当前 socket 接收消息
func (s *NetlinkSocket) Receive() ([]syscall.NetlinkMessage, error) {
	// 获取分页的页面大小（page size），这里分一页（一般 4k），是为了效率最高
	pageSize := syscall.Getpagesize()
	// 创建缓冲区，一个 page size 大小
	receiveBuff := make([]byte, pageSize)

	// 从 Socket 中接收返回结果，保存在缓冲区receiveBuff中
	lenBuff, _, err := syscall.Recvfrom(s.fd, receiveBuff, 0)
	if err != nil {
		return nil, err
	}

	// 检查数据长度
	if lenBuff < syscall.NLMSG_HDRLEN { // 如果接收数据的总长度还不够 NLMSG 的header长度
		return nil, ErrShortResponse
	}

	// 取出数据，转成 []NetlinkMessage 结构
	receiveBuff = receiveBuff[:lenBuff]
	return syscall.ParseNetlinkMessage(receiveBuff)
}

// GetPid 用于解析 socket fd，从中提取出与特定进程(端口)的 PID。
func (s *NetlinkSocket) GetPid() (uint32, error) {
	// 从套接字中获取 pid
	sa, err := syscall.Getsockname(s.fd)
	if err != nil {
		return 0, err
	}

	switch v := sa.(type) {
	case *syscall.SockaddrNetlink:
		return v.Pid, nil
	}

	return 0, ErrWrongSockType
}

// CheckMessage 校验消息的合法性，校验消息中的 seq 和 pid 是否与参数一致
// seq - 消息序列号，用以将消息排队\
// pid - 进程（端口）ID 号，用户进程来说就是其 socket 所绑定的 ID 号
func (s *NetlinkSocket) CheckMessage(msg syscall.NetlinkMessage, seq, pid uint32) error {
	// 检查返回的消息的 Header 中的 Seq 是否一致
	if msg.Header.Seq != seq {
		return fmt.Errorf("netlink -: invalid seq %d, expected %d", msg.Header.Seq, seq)
	}

	// 检查返回的消息的 Header 中的 Pid 是否一致
	if msg.Header.Pid != pid {
		return fmt.Errorf("netlink -: wrong pid %d, expected %d", msg.Header.Pid, pid)
	}

	switch msg.Header.Type {
	case syscall.NLMSG_DONE: // 如果返回消息的类型是 Done，表示已经接收完成
		return io.EOF
	case syscall.NLMSG_ERROR: // 如果返回消息的类型是 ERROR，获取错误码, data前4位表示长度
		errNum := int64(native.Uint32(msg.Data[0:4]))
		if errNum == 0 {
			return io.EOF
		}
		return syscall.Errno(-errNum)
	}
	return nil
}

// HandleAck 接收并校验消息的完整性
func (s *NetlinkSocket) HandleAck(seq uint32) error {
	pid, err := s.GetPid()
	if err != nil {
		return err
	}

outer:
	for {
		msgs, err := s.Receive() // 接收所有消息体
		if err != nil {
			return err
		}
		for _, msg := range msgs { // 遍历所有消息，校验
			if err := s.CheckMessage(msg, seq, pid); err != nil {
				if err == io.EOF { // 如果消息以 EOF 结束，表示接收成功
					break outer // 跳出外层循环
				}
				return err
			}

		}
	}
	return nil
}
