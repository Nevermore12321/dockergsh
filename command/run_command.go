package command

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"

	"github.com/Nevermore12321/dockergsh/cmdExec"
	"github.com/Nevermore12321/dockergsh/container"
)

// 定义了 docker run 命令的 RunCommand 的所有 Flags，也就是用 -- 来指定的选项
var RunCommand = &cli.Command{
	Name: "run",
	Usage: `Create a container with namespace and cgroup limit
			mydocker run -it [command]`,
	Flags: []cli.Flag{
		&cli.BoolFlag{							// docker run -it 命令
			Name: "it",
			Usage: "enable tty",
		},
	},
	/*
	这里是run命令执行的真正函数。
	1.判断参数是否包含 command
	2.获取用户指定的 command
	3.调用 Runfunction 去准备启动容器
	*/
	Action: func(context *cli.Context) error {
		if context.NArg() < 1 {
			return fmt.Errorf("Missing container command")
		}

		var cmdArray []string

		for i := 0; i < context.NArg(); i++ {
			cmdArray = append(cmdArray, context.Args().Get(i))
		}

		// docker run --it [imageName]

		imageName := cmdArray[0]
		cmdArray = cmdArray[1:]
		fmt.Println(imageName)

		// -it 和 -d 不能同时使用
		tty := context.Bool("it")
		detach := context.Bool("d")

		if tty && detach {
			return fmt.Errorf("-it and -d paramter can not both provided")
		}

		cmdExec.Run(tty, cmdArray)

		return nil
	},

}

// 定义了 InitCommand 的具体操作，此操作为内部方法，禁止外部调用
// 其实就是 container 内有一个 init 进程
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