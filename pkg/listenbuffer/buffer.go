package listenbuffer

import "net"

// NewListenBuffer 让Docker Server立即监听指定协议地址上的请求，但是将这些请求暂时先缓存下来，等Docker Daemon全部启动完毕之后，才让Docker Server开始接受这些请求
// 可以保证在Docker Daemon还没有完全启动完毕之前，接收并缓存尽可能多的用户请求。
func NewListenBuffer(proto, addr string, activate chan struct{}) (net.Listener, error) {
	wrapped, err := net.Listen(proto, addr)
	if err != nil {
		return nil, err
	}

	return defaultListener{
		wrapped:  wrapped,
		activate: activate,
	}, nil
}

type defaultListener struct {
	wrapped  net.Listener
	ready    bool
	activate chan struct{}
}

func (l defaultListener) Accept() (net.Conn, error) {
	// 如果 listener 已经准备好，开始接收请求
	if l.ready {
		return l.wrapped.Accept()
	}

	// 否则 等待 activate channel 通知
	<-l.activate
	l.ready = true
	return l.wrapped.Accept()
}

func (l defaultListener) Close() error {
	return l.wrapped.Close()
}

func (l defaultListener) Addr() net.Addr {
	return l.wrapped.Addr()
}
