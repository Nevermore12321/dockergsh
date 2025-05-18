package systemd

import (
	"errors"
	"github.com/coreos/go-systemd/activation"
	"net"
	"strconv"
)

// ListenFd socket 激活允许 systemd 直接将连接传递给应用程序，从而避免了应用程序自身监听端口的开销。
func ListenFd(addr string) ([]net.Listener, error) {
	// 返回一个由 systemd 激活的所有 socket 监听器组成的列表
	listeners, err := activation.Listeners()
	if err != nil {
		return nil, err
	}

	if listeners == nil || len(listeners) == 0 {
		return nil, errors.New("no sockets found")
	}

	// 默认允许所有 unix:// and tcp://
	if addr == "" {
		addr = "*"
	}

	// 因为有标准输入、输出、错误，所以从 3 开始
	fdNum, _ := strconv.Atoi(addr)
	fdOffset := fdNum - 3
	if addr != "*" && (len(listeners) < fdOffset+1) {
		return nil, errors.New("not enough sockets found")
	}
	if addr == "*" {
		return listeners, nil
	}
	return []net.Listener{listeners[fdOffset]}, nil
}
