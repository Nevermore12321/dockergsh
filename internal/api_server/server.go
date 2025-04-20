package api_server

import (
	"github.com/Nevermore12321/dockergsh/internal/engine"
	log "github.com/sirupsen/logrus"
	"net"
	"strings"
)

var (
	activationLock chan struct{}
)

// ServeApi dockergsh api-server 启动，args 是监听的 host 地址
func ServeApi(job *engine.Job) engine.Status {
	// 如果没有传 host 参数，报错
	if len(job.Args) == 0 {
		return job.Errorf("usage: %s PROTO://ADDR [PROTO://ADDR ...]", job.Name)
	}

	var (
		protoAddrs = job.Args                          // 传入参数
		chError    = make(chan error, len(protoAddrs)) // 返回错误 channal
	)
	activationLock = make(chan struct{}) // 启动成功 channal

	// 新建 goroutine 启动监听程序
	for _, addr := range protoAddrs {
		protoAddrParts := strings.SplitN(addr, "://", 2)
		if len(protoAddrParts) != 2 {
			return job.Errorf("invalid address format: %s, usage: %s PROTO://ADDR [PROTO://ADDR ...]", addr, job.Name)
		}
		go func() {
			log.Info("Listening for HTTP on %s (%s)", protoAddrParts[0], protoAddrParts[1])
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

func ListenAndServe(proto, addr string, job *engine.Job) error {
	var listener net.Listener
	router, err := createRouter()
	if err != nil {
		return err
	}
}
