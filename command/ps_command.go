package command

import (
	"github.com/Nevermore12321/dockergsh/cmdExec"
	"github.com/urfave/cli/v2"
)

var ListCommand = &cli.Command{
	Name: "ps",
	Usage: "list all the containers",
	Action: func(context *cli.Context) error {
		cmdExec.ListContainers()
		return nil
	},
}
