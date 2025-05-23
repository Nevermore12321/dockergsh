package api_server

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/Nevermore12321/dockergsh/internal/engine"
	"github.com/Nevermore12321/dockergsh/pkg/listenbuffer"
	"github.com/Nevermore12321/dockergsh/pkg/systemd"
	log "github.com/sirupsen/logrus"
	"net"
	"net/http"
	"os"
	"strings"
	"syscall"
)

var (
	activationLock chan struct{}
)

// Docker Server支持的协议包括以下三种：TCP协议、UNIX Socket形式以及fd的形式

// ServeFd 通过文件启动 socker 监听服务
func ServeFd(addr string, handler http.Handler) error {
	listeners, err := systemd.ListenFd(addr)
	if err != nil {
		return err
	}

	// 所有需要监听 socket goroutine 错误信息
	chErrors := make(chan error, len(listeners))

	// 在 daemon 还没有初始化、安装之前，并不想启动监听服务
	// 阻塞等待
	<-activationLock

	// 如果 listener 返回多个，说明需要监听多个 socket，启动 goroutine
	for i := range listeners {
		listener := listeners[i]
		go func() {
			log.Info("Listening for HTTP on fd %s", listener.Addr().String())
			httpSrv := http.Server{Handler: handler}
			chErrors <- httpSrv.Serve(listener)
		}()
	}

	// 等待所有 goroutine 结束
	for i := 0; i < len(listeners); i++ {
		err := <-chErrors
		if err != nil {
			return err
		}
	}
	return nil
}

// ServeApi dockergsh api-server 启动，args 是监听的 host 地址
func ServeApi(job *engine.Job) engine.Status {
	// 如果没有传 host 参数，报错
	if len(job.Args) == 0 {
		return job.Errorf("usage: %s PROTO://ADDR [PROTO://ADDR ...]", job.Name)
	}

	var (
		protoAddrs = job.Args                          // 传入参数
		chError    = make(chan error, len(protoAddrs)) // 返回错误 channal，开启了几种协议的 server，就需要有多长的错误 channel
	)

	// serveapi运行时，ServeFd和ListenAndServe函数均由于activationLock中没有内容而阻塞，
	// 而当运行acceptionconnections这个Job时，该Job会首先通知init进程Docker Daemon已经启动完毕，
	// 并关闭activationLock，因此ServeFd以及ListenAndServe不再阻塞
	activationLock = make(chan struct{}) // 用以同步 serveapi 和 acceptconnections 这两个Job执行的管道

	// 新建 goroutine 启动监听程序
	for _, addr := range protoAddrs {
		protoAddrParts := strings.SplitN(addr, "://", 2)
		if len(protoAddrParts) != 2 {
			return job.Errorf("invalid address format: %s, usage: %s PROTO://ADDR [PROTO://ADDR ...]", addr, job.Name)
		}
		go func() {
			log.Infof("Listening for HTTP on %s (%s)", protoAddrParts[0], protoAddrParts[1])
			chError <- ListenAndServe(protoAddrParts[0], protoAddrParts[1], job)
		}()
	}

	for i := 0; i < len(protoAddrs); i++ {
		err := <-chError
		if err != nil {
			return job.Error(err)
		}
	}

	return engine.StatusOk
}

// ListenAndServe 基于特定协议（tcp、unix）启动 HTTP or HTTPS 服务.
func ListenAndServe(proto, addr string, job *engine.Job) error {
	var listener net.Listener
	router, err := createRouter(job.Eng, job.GetEnvBool("Logging"), job.GetEnvBool("EnableCors"), job.GetEnv("Version"))
	if err != nil {
		return err
	}

	// 根据监听的协议类型启动 server
	if proto == "fd" {
		// 若协议为fd形式，则直接通过ServeFd来服务请求
		return ServeFd(addr, router)
	}

	var oldmask int // 文件打开的默认权限
	if proto == "unix" {
		// 若协议为unix形式，则先将软连接删掉
		if err := syscall.Unlink(addr); err != nil && !os.IsNotExist(err) {
			return err
		}
		oldmask = syscall.Umask(0777)
	}

	// 如果需要缓存 request
	if job.GetEnvBool("BufferRequests") {
		listener, err = listenbuffer.NewListenBuffer(proto, addr, activationLock)
	} else {
		listener, err = net.Listen(proto, addr)
	}

	if proto == "unix" {
		syscall.Umask(oldmask)
	}

	if err != nil {
		return err
	}

	// Job中环境变量Tls或者TlsVerify有一个为true，则说明Docker Server需要支持HTTPS服务
	if proto != "unix" && (job.GetEnvBool("Tls") || job.GetEnvBool("TlsVerify")) {
		// 获取相关证书、密钥文件
		tlsCert := job.GetEnv("TlsCert")
		tlsKey := job.GetEnv("TlsKey")

		// 加载证书
		cert, err := tls.LoadX509KeyPair(tlsCert, tlsKey)
		if err != nil {
			return fmt.Errorf("couldn't load X509 key pair (%s, %s): %s. Key encrypted",
				tlsCert, tlsKey, err)
		}

		var tlsConfig = &tls.Config{
			NextProtos:   []string{"h2", "http/1.1"},
			Certificates: []tls.Certificate{cert},
		}

		// 校验客户端证书
		if job.GetEnvBool("TlsVerify") {
			certPool := x509.NewCertPool()
			caFile, err := os.ReadFile(job.GetEnv("TlsCa"))
			if err != nil {
				return fmt.Errorf("couldn't read CA certificate: %s", err)
			}
			certPool.AppendCertsFromPEM(caFile)

			tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
			tlsConfig.ClientCAs = certPool
		}

		listener = tls.NewListener(listener, tlsConfig)
	}

	// 参数校验
	switch proto {
	case "tcp":
		if !strings.HasPrefix(addr, "127.0.0.1") && job.GetEnvBool("TlsVerify") {
			log.Infof("/!\\ DON'T BIND ON ANOTHER IP ADDRESS THAN 127.0.0.1 IF YOU DON'T KNOW WHAT YOU'RE DOING /!\\")
		}
	case "unix":
		socketGroup := job.GetEnv("SocketGroup")
		if socketGroup != "" {
			if err := changeGroup(addr, socketGroup); err != nil {
				return err
			} else {
				log.Debugf("Warning: couldn't change group of socket %s to %s.", addr, socketGroup)
			}
		}
		if err := os.Chmod(addr, 0660); err != nil {
			return err
		}
	default:
		return fmt.Errorf("invalid protocol format")
	}

	httpSrv := http.Server{
		Addr:    addr,
		Handler: router,
	}
	return httpSrv.Serve(listener)
}

func AcceptConnections(job *engine.Job) engine.Status {
	// 发送 ready 信号
	//go func() {
	//	err := systemd.SendNotify("READY=1")
	//	if err != nil {
	//		log.Errorf("failed to send READY=1: %v", err)
	//	}
	//}()

	// 如果 activation channel 不为空， 关闭，表示已经可以接收请求
	if activationLock != nil {
		close(activationLock)
	}
	return engine.StatusOk
}
