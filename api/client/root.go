package client

import (
	"github.com/Nevermore12321/dockergsh/api"
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
	RootCmd.Flags = api.CmdFlags()

	RootCmd.Action = rootAction
	RootCmd.Before = api.RootBefore
	RootCmd.After = rootAfter

	// 初始化子命令行
	InitSubCmd(RootCmd)
}

func rootAction(context *cli.Context) error {
	if err := api.PreCheckConfDebug(context); err != nil {
		return err
	}

	protohost, err := api.PreCheckConfHost(context)
	if err != nil {
		return err
	}

	tlsConfig, err := api.PreCheckConfTLS(context)
	if err != nil {
		return err
	}

	// 初始化 dockergshclient
	// 创建Docker Client实例。
	DockerGshCliInitial(context.App.Reader, context.App.Writer, context.App.ErrWriter, protohost[0], protohost[1], tlsConfig)

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
