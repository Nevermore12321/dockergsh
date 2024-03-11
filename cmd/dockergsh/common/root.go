package common

import (
	"github.com/Nevermore12321/dockergsh/cmd/dockergsh/subcmds"
	"github.com/Nevermore12321/dockergsh/internal/client"
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
	RootCmd.Usage = GetHelpUsage("")

	// 初始化版本
	RootCmd.Version = VERSION

	// 初始化 RootCmd 的 flags
	RootCmd.Flags = cmdCommonFlags()

	RootCmd.Action = rootAction
	RootCmd.Before = rootBefore
	RootCmd.After = rootAfter

	// 初始化子命令行
	subcmds.InitSubCmd(RootCmd)
}

func rootAction(context *cli.Context) error {
	if err := PreCheckConfDebug(context); err != nil {
		return err
	}

	protohost, err := PreCheckConfHost(context)
	if err != nil {
		return err
	}

	tlsConfig, err := PreCheckConfTLS(context)
	if err != nil {
		return err
	}

	// 初始化 dockergshclient
	// 创建Docker Client实例。
	client.DockergshCliInitial(context.App.Reader, context.App.Writer, context.App.ErrWriter, protohost[0], protohost[1], tlsConfig)

	return nil
}

func rootBefore(context *cli.Context) error {
	// 命令运行前的初始化 logrus 的日志配置
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetOutput(context.App.Writer)
	return nil
}

func rootAfter(context *cli.Context) error {
	err := context.Err()
	if err != nil {
		logrus.Info(err)
	}
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
