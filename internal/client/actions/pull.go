package actions

import (
	"errors"
	"github.com/urfave/cli/v2"
)

var (
	ErrCmdFormat = errors.New("the format of the command you entered is incorrect. Please use -h to check usage")
)

func CmdPull(context *cli.Context) error {
	if context.NArg() != 1 {
		return ErrCmdFormat
	}
	return nil
}
