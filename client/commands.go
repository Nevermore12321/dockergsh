package client

import (
	"errors"
	"github.com/urfave/cli/v2"
)

var (
	ErrCmdFormat = errors.New("The format of the command you entered is incorrect. Please use - h to check usage")
)

func InitSubCmd(root *cli.App) {
	root.Commands = cli.Commands{
		&cli.Command{
			Name:   "pull",
			Usage:  GetHelpUsage("pull"),
			Action: CmdPull,
		},
	}

}

func CmdPull(context *cli.Context) error {
	if context.NArg() != 1 {
		return ErrCmdFormat
	}
	return nil
}
