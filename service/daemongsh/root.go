package daemongsh

import (
	"github.com/Nevermore12321/dockergsh/service"
	"github.com/Nevermore12321/dockergsh/service/client"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"io"
)

var (
	RootCmd = cli.NewApp()
)

func RootCmdInitial(name string, in io.Reader, out, err io.Writer) {
	RootCmd.Name = name
	if in != nil {
		RootCmd.Reader = in
	}
	if out != nil {
		RootCmd.Writer = out
	}
	if err != nil {
		RootCmd.ErrWriter = err
	}

	// help 信息
	RootCmd.Usage = client.GetHelpUsage("")

	// 初始化版本
	RootCmd.Version = client.VERSION

	// 初始化 RootCmd 的 flags
	// 添加 daemongsh 的 flags
	RootCmd.Flags = append(service.CmdFlags(), daemongshFlags()...)

	RootCmd.Action = rootAction
	RootCmd.Before = service.RootBefore
	RootCmd.After = rootAfter
}

func rootAction(context *cli.Context) error {
	if err := service.PreCheckConfDebug(context); err != nil {
		return err
	}

	protohost, err := service.PreCheckConfHost(context)
	if err != nil {
		return err
	}

	tlsConfig, err := service.PreCheckConfTLS(context)
	if err != nil {
		return err
	}

	// 初始化 dockergshclient
	// 创建Docker Client实例。
	client.DockerGshCliInitial(context.App.Reader, context.App.Writer, context.App.ErrWriter, protohost[0], protohost[1], tlsConfig)

	return nil
}

func rootAfter(context *cli.Context) error {
	err := context.Err()
	logrus.Info(err)
	// todo http status error
	//if err != nil {
	//	if sterr, ok := err.(*StatusError); ok {
	//		if sterr.Status != "" {
	//			log.Println(sterr.Status)
	//		}
	//		os.Exit(sterr.StatusCode)
	//	}
	//}
	return nil
}

func daemongshFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:    "pidfile",
			Aliases: []string{"p"},
			Value:   "/var/run/docker.pid",
			Usage:   "Path to use for daemon PID file",
		},
	}
}

// daemongsh 的启动流程
func mainDaemon(context *cli.Context) {
	//1）daemon的配置初始化。这部分在init（）函数中实现，即在mainDaemon（）运行前就执行，但由于这部分内容和mainDaemon（）的运行息息相关，可以认为是mainDaemon（）运行的先决条件。
	//2）命令行flag参数检查。
	//3）创建engine对象。
	//4）设置engine的信号捕获及处理方法。
	//5）加载builtins。
	//6）使用goroutine加载daemon对象并运行。
	//7）打印Docker版本及驱动信息。
	//8）serveapi的创建与运行。
}
