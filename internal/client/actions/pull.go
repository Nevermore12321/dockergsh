package actions

import (
	"errors"
	"fmt"
	"github.com/Nevermore12321/dockergsh/internal/utils"
	"github.com/urfave/cli/v2"
	"os"
)

var (
	ErrCmdFormat = errors.New("the format of the command you entered is incorrect. Please use -h to check usage")
)

func CmdPull(context *cli.Context) error {
	if context.NArg() != 1 {
		return ErrCmdFormat
	}
	fmt.Println(os.Getenv(utils.DockergshHosts))
	return nil
}
