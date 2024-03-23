package client

import (
	"crypto/tls"
	"github.com/Nevermore12321/dockergsh/internal/registry"
	"github.com/Nevermore12321/dockergsh/pkg/terminal"
	"io"
	"os"
)

// DockerGshClient dockergsh client
type DockerGshClient struct {
	proto      string               // c/s 之间通信的协议，unix，tcp，fd
	addr       string               // server 端的地址
	scheme     string               // http or https
	tlsConfig  *tls.Config          // tls 配置
	in         io.Reader            // input interface
	out        io.Writer            // out interface
	err        io.Writer            // error interface
	configFile *registry.ConfigFile // 仓库配置信息
	isTerminal bool                 // 终端模式开关
	terminalFd uintptr              // 终端模式文件句柄
}

func NewDockergshClient(in io.Reader, out, err io.Writer, proto, addr string, tlsConf *tls.Config) *DockerGshClient {
	// 默认值
	var (
		scheme     = "http" // 默认使用 http 协议
		isTerminal = false  // 默认不开启终端模式
		terminalFd uintptr  // 默认终端文件句柄为空
	)

	// 如果配置了 tls，那么就是用 https
	if tlsConf != nil {
		scheme = "https"
	}

	// in 不为空，同时输出 out 可以转化为文件类型，那么获取文件句柄，同时判断是否为终端类型
	if in != nil {
		// 如果 out 也是文件格式
		if file, ok := out.(*os.File); ok {
			terminalFd = file.Fd()                       // 获取out文件句柄给 terminalFd
			isTerminal = terminal.IsTerminal(terminalFd) // 判断是否为终端类型
		}
	}

	// 如果没有指定错误输出，那么输出作为错误输出。
	if err == nil {
		err = out
	}
	return &DockerGshClient{
		proto:      proto,
		addr:       addr,
		in:         in,
		out:        out,
		err:        err,
		isTerminal: isTerminal,
		terminalFd: terminalFd,
		tlsConfig:  tlsConf,
		scheme:     scheme,
	}

}
