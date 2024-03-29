package command

import (
	"fmt"
	"github.com/Nevermore12321/dockergsh/cgroup/subsystem"
	"github.com/urfave/cli/v2"

	"github.com/Nevermore12321/dockergsh/cmdExec"
)

// 定义了 docker run 命令的 RunCommand 的所有 Flags，也就是用 -- 来指定的选项
var RunCommand = &cli.Command{
	Name: "run",
	Usage: `Create a container with namespace and cgroup limit
			mydocker run -it [command]`,
	Flags: []cli.Flag{
		&cli.BoolFlag{ // docker run -it 命令
			Name:  "it",
			Usage: "enable tty",
		},
		&cli.BoolFlag{
			Name:  "d",
			Usage: "detach container",
		},
		&cli.StringFlag{
			Name:  "m",
			Usage: "memory limit",
		},
		&cli.StringFlag{
			Name:  "cpu",
			Usage: "cpu limit",
		},
		&cli.StringFlag{
			Name:  "cpuset",
			Usage: "cpuset limit",
		},
		&cli.StringFlag{
			Name:  "v",
			Usage: "Volume",
		},
		&cli.StringFlag{
			Name:  "name",
			Usage: "container name",
		},
		&cli.StringSliceFlag{
			Name:  "e",
			Usage: "set environments",
		},
		&cli.StringFlag{
			Name:  "net",
			Usage: "container network",
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

		// 要执行的 命令
		var cmdArray []string
		for i := 0; i < context.NArg(); i++ {
			cmdArray = append(cmdArray, context.Args().Get(i))
		}

		// docker run --it [imageName]
		// docker run -it -m 100m busybox stress --vm-bytes 200m --vm-keep -m 1

		// 获取 image name
		imageName := cmdArray[0]
		cmdArray = cmdArray[1:]

		// 获取 选项参数变量
		// volume
		volume := context.String("v")

		// 获取 container network 参数变量
		network := context.String("net")

		// container name
		containerName := context.String("name")

		// 环境变量
		envSlice := context.StringSlice("e")

		// -it 和 -d 不能同时使用
		tty := context.Bool("it")
		detach := context.Bool("d")

		if tty && detach {
			return fmt.Errorf("-it and -d paramter can not both provided")
		}

		// cgroup 资源配置
		resConf := &subsystem.ResourceConfig{
			MemoryLimit: context.String("m"),
			CpuShare:    context.String("cpu"),
			CpuSet:      context.String("cpuset"),
		}

		cmdExec.Run(tty, cmdArray, resConf, imageName, containerName, volume, envSlice, network)

		return nil
	},
}
