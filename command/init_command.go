package command

import (
	"github.com/Nevermore12321/dockergsh/container"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

// 定义了 InitCommand 的具体操作，此操作为内部方法，禁止外部调用
// 其实就是 container 内有一个 init 进程
// init 命令的作用：fork出子进程启动容器时，执行 /proc/self/exe 来执行 dockergsh init 命令来启动容器
var InitCommand = &cli.Command{
	Name: "init",
	Usage: "Init container process run user's process in container. Do not call it outside",
	/*
		1. 获取传递过来的 command 参数
		2. 执行容器初始化操作
	*/
	Action: func(context *cli.Context) error {
		log.Infof("init come on")
		cmd := context.Args().Get(0)
		log.Infof("Command %s", cmd)
		err := container.RunContainerInitProcess()
		return err
	},
}