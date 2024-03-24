package builtins

/*
builtins 是 Docker Daemon运行过程中，注册的一些任务（Jobs）
	这部分任务一般与容器的运行无关，与Docker Daemon的运行时信息有关
注册的任务不会立即执行，当Docker Daemon接收到Job的执行请求时，才被Docker Daemon调用执行
*/

import (
	"github.com/Nevermore12321/dockergsh/internal/engine"
	"github.com/Nevermore12321/dockergsh/internal/events"
	"github.com/Nevermore12321/dockergsh/internal/registry"
)

// Register 加载 builtins，向 engine 注册多个 Handler，以便后续在执行相应任务时，运行指定的 Handler
// 这些Handler包括：Docker Daemon宿主机的网络初始化、Web API服务、事件查询、版本查看、Docker Registry的验证与搜索等
func Register(eng *engine.Engine) error {
	// 1. 注册网络初始化处理方法
	if err := daemon(eng); err != nil {
		return err
	}

	// 2. 注册API服务处理方法
	if err := remote(eng); err != nil {
		return err
	}

	// 3. 注册events事件处理方法
	if err := events.New().Install(eng); err != nil {
		return err
	}

	// 4. 注册版本处理方法
	if err := eng.Register("version", dockergshVersion); err != nil {
		return err
	}

	// 5. 注册registry(镜像仓库)处理方法
	return registry.NewService().Install(eng)
}

// 网络初始化 job，todo 网络栈初始化
// 1. 获取为Docker服务的网络设备地址。
// 2. 创建指定IP地址的网桥。
// 3. 配置网络iptables规则。
// 4. 另外还为 eng 对象注册了多个 Handler，如 allocate_interface、release_interface、allocate_port以及link等。
func daemon(eng *engine.Engine) error {
	return eng.Register("init_networkdriver", bridge.InitDriver)
}

// API 初始化 job
// 1. ServeApi执行时，通过循环多种指定协议，创建出goroutine来配置指定的http.Server，最终为不同协议的请求服务（也就是 server 接收请求）
// 2. AcceptConnections的作用主要是：通知宿主机上init守护进程Docker Daemon已经启动完毕，可以让Docker Daemon开始服务API请求
func remote(eng *engine.Engine) error {
	if err := eng.Register("serveapi", apiserver.ServeApi); err != nil {
		return err
	}

	return eng.Register("acceptconnections", apiserver.AcceptConnections)
}

// 名为 `version` Job 的处理函数 handler
// dockergshVersion 会向名为version的Job的标准输出中写入:
// Docker的版本、Docker API的版本、git版本、Go语言运行时版本，以及操作系统版本等信息
func dockergshVersion(job *engine.Job) engine.Status {
	// todo
	return 0
}
