package client

import (
	"crypto/tls"
	"io"
)

type DockerGshClient struct {
	proto     string      // c/s 之间通信的协议，unix，tcp，fd
	addr      string      // server 端的地址
	scheme    string      // http or https
	tlsConfig *tls.Config // tls 配置
	in        io.Reader   // input interface
	out       io.Writer   // out interface
	err       io.Writer   // error interface
}

var DockerGshCli *DockerGshClient = new(DockerGshClient)

func DockergshCliInitial(in io.Reader, out, err io.Writer, proto, addr string, tlsConf *tls.Config) {
	// 默认值
	var (
		scheme = "http"
	)

	// 如果配置了 tls，那么就是用 https
	if tlsConf != nil {
		scheme = "https"
	}

	// todo in 可以为 file 文件格式

	// 如果没有指定错误输出，那么输出作为错误输出。
	if err == nil {
		err = out
	}

	DockerGshCli.proto = proto
	DockerGshCli.addr = addr
	DockerGshCli.scheme = scheme
	DockerGshCli.tlsConfig = tlsConf
	DockerGshCli.in = in
	DockerGshCli.out = out
	DockerGshCli.err = err

}
