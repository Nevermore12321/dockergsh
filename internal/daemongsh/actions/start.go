package actions

import (
	"errors"
	"github.com/Nevermore12321/dockergsh/internal/builtins"
	"github.com/Nevermore12321/dockergsh/internal/daemongsh"
	"github.com/Nevermore12321/dockergsh/internal/engine"
	"github.com/Nevermore12321/dockergsh/internal/utils"
	"github.com/Nevermore12321/dockergsh/pkg/signal"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

var (
	daemonCfg = &daemongsh.Config{}
)

var (
	ErrCmdFormat = errors.New("the format of the command you entered is incorrect. Please use -h to check usage")
)

func CmdStart(context *cli.Context) error {
	if context.NArg() != 1 {
		return ErrCmdFormat
	}
	// todo parse hosts
	mainDaemon(context, nil)
	return nil
}

// mainDaemon daemon 的启动流程
func mainDaemon(context *cli.Context, hosts []string) {
	// 1.daemon的配置初始化
	daemonCfg.InitialFlags(context)

	//2. 命令行flag参数个数检查。
	if context.NArg() != 0 {
		return
	}
	//3. 创建engine对象。
	eng := engine.New()

	//4）设置engine的信号捕获及处理方法。
	signal.Trap(eng.Shutdown)

	//5）加载builtins。给 engine 注册不同的任务job
	if err := builtins.Register(eng); err != nil {
		log.Fatal(err)
	}

	//6）使用goroutine加载daemon对象并运行。
	go func() {
		// 6.1 创建daemon对象
		// 初始化基本环境，如处理config参数，验证系统支持度，配置Docker工作目录，设置与加载多种驱动，创建graph环境，验证DNS配置等
		daemon, err := daemongsh.NewDaemongsh(daemonCfg, eng)
		if err != nil {
			log.Fatal(err)
		}

		// 6.2 通过daemon对象为engine注册Handler
		if err := daemon.Install(eng); err != nil {
			log.Fatal(err)
		}

		// 6.3 创建名为acceptconnections的Job，并且开始运行
		if err := eng.Job("acceptconnections").Run(); err != nil {
			log.Fatal(err)
		}
	}()

	//7）打印Docker版本及驱动信息。
	log.Printf("docker daemon: %s %s; execdriver: %s; graphdriver: %s",
		utils.VERSION,
		utils.GITCOMMIT,
		daemonCfg.ExecDriver,
		daemonCfg.GraphDriver,
	)

	//8）serveapi的创建与运行。
	// 8.1 hosts 为 Dockergsh Daemon提供使用的协议与监听的地址
	job := eng.Job("serveapi", hosts...)
	// 8.2 设置环境变量
	job.SetEnvBool(utils.Logging, true)
	job.SetEnvBool(utils.EnableCors, context.Bool("api-enable-cors"))
	job.SetEnv(utils.Version, utils.VERSION)
	job.SetEnv(utils.SocketGroup, context.String("socket-group"))

	job.SetEnvBool(utils.Tls, context.Bool("tls"))
	job.SetEnvBool(utils.TlsVerify, context.Bool("tls-verify"))
	job.SetEnv(utils.TlsCa, context.String("tls-cacert"))
	job.SetEnv(utils.TlsCert, context.String("tls-cert"))
	job.SetEnv(utils.TlsKey, context.String("tls-key"))
	job.SetEnvBool(utils.BufferRequests, true)

	// 8.3 运行 job
	if err := job.Run(); err != nil {
		log.Fatal(err)
	}
}
