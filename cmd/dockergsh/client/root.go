package client

import (
	gshClient "github.com/Nevermore12321/dockergsh/client"
	service "github.com/Nevermore12321/dockergsh/cmd"
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
	RootCmd.Usage = gshClient.GetHelpUsage("")

	// 初始化版本
	RootCmd.Version = gshClient.VERSION

	// 初始化 RootCmd 的 flags
	RootCmd.Flags = service.CmdFlags()

	RootCmd.Action = rootAction
	RootCmd.Before = service.RootBefore
	RootCmd.After = rootAfter

	// 初始化子命令行
	gshClient.InitSubCmd(RootCmd)
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
	gshClient.DockerGshCliInitial(context.App.Reader, context.App.Writer, context.App.ErrWriter, protohost[0], protohost[1], tlsConfig)

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
