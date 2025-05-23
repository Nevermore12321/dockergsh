package root

import (
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"io"
)

var (
	RootCmd = cli.NewApp()
)

func Initial(name string, in io.Reader, out, err io.Writer) {
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

	// 自动补全
	RootCmd.EnableBashCompletion = true

	// help 信息
	RootCmd.Usage = GetHelpUsage("")

	// 初始化版本
	RootCmd.Version = VERSION

	// 初始化 RootCmd 的 flags
	RootCmd.Flags = cmdCommonFlags()

	RootCmd.Before = rootBefore
	RootCmd.After = rootAfter

	// 初始化子命令行
	initSubCmd(RootCmd)
}

func rootBefore(context *cli.Context) error {
	// 命令运行前的初始化 logrus 的日志配置
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetOutput(context.App.Writer)

	// todo reexec

	// 是否开启 debug
	if err := PreCheckConfDebug(context); err != nil {
		return err
	}

	// 检查 host 是否合法
	if err := PreCheckConfHost(context); err != nil {
		return err
	}
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
