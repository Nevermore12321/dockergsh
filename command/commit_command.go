package command

import (
	"fmt"
	"github.com/Nevermore12321/dockergsh/cmdExec"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

var CommitCommand = &cli.Command{
	Name:  "commit",
	Usage: "commit a container into images",
	Action: func(context *cli.Context) error {
		if context.NArg() < 2 {
			return fmt.Errorf("missing container name and image name")
		}
		//  docker commit container image
		// 获取 container 名称，需要打包成的 image 名称
		containerArg := context.Args().Get(0)
		imageName := context.Args().Get(1)

		err := cmdExec.CommitContainer(containerArg, imageName)
		if err != nil {
			log.Errorf("commit container err: %v", err)
			return err
		}
		return nil
	},
}
