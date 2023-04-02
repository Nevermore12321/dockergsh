package command

import (
	"fmt"
	"github.com/Nevermore12321/dockergsh/cmdExec"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

var StopCommand = &cli.Command{
	Name:  "stop",
	Usage: "Stop one or more running containers",
	Action: func(context *cli.Context) error {
		// dockergsh stop [containerName or containerId]
		if context.NArg() < 1 {
			return fmt.Errorf("missing container name")
		}
		containerArg := context.Args().Get(0)
		err := cmdExec.StopContainer(containerArg)
		if err != nil {
			log.Errorf("Stop Container failed %v", err)
			return err
		}
		return nil
	},
}
