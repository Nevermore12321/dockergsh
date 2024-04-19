package actions

import (
	"errors"
	"flag"
	"fmt"
	"github.com/Nevermore12321/dockergsh/internal/builtins"
	"github.com/Nevermore12321/dockergsh/internal/daemongsh"
	"github.com/Nevermore12321/dockergsh/internal/engine"
	"github.com/Nevermore12321/dockergsh/internal/utils"
	"github.com/Nevermore12321/dockergsh/pkg/signal"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"os"
	"strings"
)

var (
	daemonCfg = &daemongsh.Config{}
)

var (
	ErrCmdFormat = errors.New("the format of the command you entered is incorrect. Please use -h to check usage")
)

func CmdStart(context *cli.Context) error {
	if context.NArg() != 0 {
		return ErrCmdFormat
	}
	// 从环境变量中获取 hosts 信息
	hostsEnv := os.Getenv(utils.DockergshHosts)
	hosts := strings.Split(hostsEnv, ",")

	mainDaemon(context, hosts)
	return nil
}

// mainDaemon daemon 的启动流程
func mainDaemon(context *cli.Context, hosts []string) {
	// 1.daemon的配置初始化
	daemonCfg.InitialFlags(context)

	//2. 命令行flag参数个数检查。
	if context.NArg() != 0 {
		daemonUsage()
		return
	}
	// 3. 创建engine对象。
	eng := engine.New()

	// 4. 设置engine的信号捕获及处理方法。
	//		因为 Daemon 是 linux 上一个后台进程，应该具有处理信号的能力
	signal.Trap(eng.Shutdown)

	// 5. 加载builtins。给 engine 注册不同的任务job
	if err := builtins.Register(eng); err != nil {
		log.Fatal(err)
	}

	// 6. 使用goroutine加载daemon对象并运行。
	//		在后台加载守护进程，即启动 http api，以便在守护进程启动时连接不会失败
	go func() {
		// 6.1 创建daemon对象
		// 	初始化基本环境，如处理config参数，验证系统支持度，配置Docker工作目录，设置与加载多种驱动，创建graph环境，验证DNS配置等
		daemon, err := daemongsh.NewDaemongsh(daemonCfg, eng)
		if err != nil {
			log.Fatal(err)
		}

		// 6.2 通过daemon对象为engine注册Handler
		if err := daemon.Install(eng); err != nil {
			log.Fatal(err)
		}

		// 6.3 创建名为 acceptconnections 的Job，并且开始运行
		// 		守护进程完成设置后，可以告诉 serveapi 开始接受连接
		// 		builtins 注册时，已经将 acceptconnections 注册到 engine 的 handler中，这里创建 Job 的 handler 就是对应在 engine 中的 handler
		if err := eng.Job("acceptconnections").Run(); err != nil {
			log.Fatal(err)
		}
	}()

	// 7）打印Docker版本及驱动信息。
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

func daemonUsage() {
	fmt.Fprintf(os.Stderr, "Usage: docker daemon start <flags>\n")
	flag.PrintDefaults()
	os.Exit(1)
}
