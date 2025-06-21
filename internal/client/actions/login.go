package actions

import (
	"fmt"
	"github.com/Nevermore12321/dockergsh/internal/client"
	"github.com/Nevermore12321/dockergsh/internal/registry"
	"github.com/urfave/cli/v2"
	"net/url"
)

func CmdLoginFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:    "username",
			Aliases: []string{"u"},
			Usage:   "username for login",
		},
		&cli.StringFlag{
			Name:    "password",
			Aliases: []string{"p"},
			Usage:   "password for login",
		},
		&cli.StringFlag{
			Name:    "email",
			Aliases: []string{"e"},
			Usage:   "email for login",
		},
	}
}

func CmdLoginDescription() string {
	return "Register or log in to a Docker registry server, if no server is specified \"" + registry.IndexServerAddress() + "\" is the default."
}

// "Docker登录"：登录注册用户到注册表服务。
func CmdLogin(context *cli.Context) error {
	username := context.String("username")
	password := context.String("password")
	email := context.String("email")

	serverAddress := registry.IndexServerAddress()
	// 如果有参数，那么就设置仓库
	if context.NArg() > 0 {
		serverAddress = context.Args().Get(0)
	}

	dockergshClient := client.Client
	// 提示用户输入字段，如果已有默认值就加上括号提示。
	promptDefault := func(prompt string, configDefault string) string {
		if configDefault == "" {
			fmt.Fprintf(dockergshClient.Out, "%s: ", prompt)
		} else {
			fmt.Fprintf(dockergshClient.Out, "%s (%s): ", prompt, configDefault)
		}
	}

	// todo
}
