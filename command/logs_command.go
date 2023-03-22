package command

import (
	"fmt"
	"github.com/Nevermore12321/dockergsh/logs"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

var LogsCommand = &cli.Command{
	Name:  "logs",
	Usage: "Fetch the logs of a container",
	Action: func(context *cli.Context) error {
		if context.NArg() < 1 {
			return fmt.Errorf("Please input your container name")
		}
		containerName := context.Args().Get(0)
		err := logs.LogContainer(containerName)
		if err != nil {
			logrus.Errorf("Log container error %v", err)
			return err
		}
		return nil
	},
}
