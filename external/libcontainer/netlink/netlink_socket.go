package netlink

import "syscall"

type NetlinkSocket struct {
	fd int
	sa syscall.SockaddrNetlink
}

// getNetlinkSocket 创建 NetlinkSocket 实例对象
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

	// 1. 创建一个 socket，协议族是 AF_NETLINK，并且使用原始套接字
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
	// 		- pid: Socket的Port号，即要发送给哪个Socket，就指定为它的Port号，默认初始化 0
	// 		- Groups: 是要发送的多播组，即将Netlink消息发送给哪些多播组，默认初始化 0
	//		- Family: 协议族，必须指定为 AF_NETLINK
	s.sa.Family = syscall.AF_NETLINK

	// 3. 将 socket 描述符与 SockaddrNetlink 绑定
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

//  netlink 就是 socket 直接的通信

// Send 发送消息给当前 socket
func (s *NetlinkSocket) Send(request *NetlinkRequest) error {
	return syscall.Sendto(s.fd, request.ToWireFormat(), 0, &s.sa)
}

// Receive 当前 socket 接收消息
func (s *NetlinkSocket) Receive() ([]syscall.NetlinkMessage, error) {

}
