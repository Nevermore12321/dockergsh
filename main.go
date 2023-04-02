package main

import (
	cmd "github.com/Nevermore12321/dockergsh/command"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"os"
)

const usage = `dockergsh is a simple container runtime implementation.
			   The purpose of this project is to learn how docker works and how to write a docker by ourselves
			   Enjoy it, just for fun.`

func main() {
	// 初始化命令行 app
	app := cli.NewApp()

	log.Println(usage)

	// 配置命令行 app
	app.Name = "dockergsh"
	app.Usage = usage
	app.Commands = []*cli.Command{
		cmd.InitCommand,
		cmd.RunCommand,
		cmd.CommitCommand,
		cmd.ListCommand,
		cmd.LogsCommand,
		cmd.ExecCommand,
		cmd.StopCommand,
	}

	// 命令运行前的初始化 logrus 的日志配置
	app.Before = func(context *cli.Context) error {
		log.SetFormatter(&log.JSONFormatter{})
		log.SetOutput(os.Stdout)
		return nil
	}

	// 执行命令行 app
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
