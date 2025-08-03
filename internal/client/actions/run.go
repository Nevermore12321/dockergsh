package actions

import (
	"fmt"
	"github.com/Nevermore12321/dockergsh/internal/utils"
	"github.com/urfave/cli/v2"
)

func CmdRunDescription() string {
	return fmt.Sprintf("Usage: docker [OPTIONS] COMMAND [arg...]\n -H=[unix://%s]: tcp://host:port to bind/connect to or unix://path/to/socket to use\n\nA self-sufficient runtime for linux containers.\n\nCommands:\n", utils.DefaultUnixSocket)
}

// CmdLogin "Docker登录"：登出
func CmdRun(context *cli.Context) error {

}
