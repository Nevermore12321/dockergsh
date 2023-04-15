package command

import (
	"fmt"
	"github.com/Nevermore12321/dockergsh/cmdExec"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

var RemoveCommand = &cli.Command{
	Name:  "rm",
	Usage: "Remove one or more containers",
	Action: func(context *cli.Context) error {
		if context.NArg() < 1 {
			return fmt.Errorf("Missing container name")
		}
		containerArg := context.Args().Get(0)
		err := cmdExec.RemoveContainer(containerArg)
		if err != nil {
			log.Errorf("Remove Container failed %v", err)
			return err
		}
		return nil
	},
}
