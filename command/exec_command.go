package command

import (
	"fmt"
	"os"

	"github.com/Nevermore12321/dockergsh/cmdExec"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

var ExecCommand = &cli.Command{
	Name:  "exec",
	Usage: "Run a command in a running container",
	Action: func(context *cli.Context) error {
		// 控制是 docker exec 第一次执行，还是添加环境变量后第二次执行 /proc/self/exe exec
		if os.Getenv("dockergsh_pid") != "" {
			// 已经添加了环境变量，第二次执行，只需要执行 C 代码即可，该文件已经导入了 C 库，因此直接返回
			log.Infof("pid callback pid %s", os.Getgid())
			return nil
		}
		if context.NArg() < 2 {
			return fmt.Errorf("Missing container name or command")
		}
		containerArg := context.Args().Get(0)
		var commandArr []string
		for _, arg := range context.Args().Tail() {
			commandArr = append(commandArr, arg)
		}

		err := cmdExec.ExecInContainer(containerArg, commandArr)
		if err != nil {
			log.Errorf("Exec Container failed %v", err)
			return err
		}
		return nil
	},
}
