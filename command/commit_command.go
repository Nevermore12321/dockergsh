package command

import (
	"fmt"
	"github.com/Nevermore12321/dockergsh/cmdExec"
	"github.com/urfave/cli/v2"
)

var CommitCommand = cli.Command{
	Name: "commit",
	Usage: "commit a container into images",
	Action: func(context *cli.Context) error {
		if context.NArg() < 1 {
			return fmt.Errorf("Missing container name")
		}
		//  docker commit container image
		// 获取 container 名称，需要打包成的 image 名称
		containerName := context.Args().Get(0)
		imageName := context.Args().Get(1)

		cmdExec.CommitContainer(containerName, imageName)
	},
}
