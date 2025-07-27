package actions

import (
	"bufio"
	"fmt"
	"github.com/Nevermore12321/dockergsh/internal/client"
	"github.com/Nevermore12321/dockergsh/internal/engine"
	"github.com/Nevermore12321/dockergsh/internal/registry"
	"github.com/Nevermore12321/dockergsh/pkg/terminal"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"io"
	"os"
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
	promptDefault := func(prompt string, configDefault string) {
		if configDefault == "" {
			fmt.Fprintf(dockergshClient.Out, "%s: ", prompt)
		} else {
			fmt.Fprintf(dockergshClient.Out, "%s (%s): ", prompt, configDefault)
		}
	}

	// 接收用户输入的用户名密码
	readInput := func(in io.Reader, out io.Writer) string {
		reader := bufio.NewReader(in)
		line, _, err := reader.ReadLine()
		if err != nil {
			fmt.Fprintln(out, err.Error())
			os.Exit(1)
		}
		return string(line)
	}

	// 加载配置文件
	err := dockergshClient.LoadConfigFile()
	if err != nil {
		log.Println("Load Config File Failed, ", err)
	}

	// 查找配置文件中，有没有当前仓库地址的配置信息
	authConfig, ok := dockergshClient.ConfigFile.Configs[serverAddress]
	if !ok {
		authConfig = registry.AuthConfig{}
	}

	// 命令行中没有指定用户名，提示用户输入用户名
	if username == "" {
		promptDefault("Username", authConfig.Username)
		username = readInput(dockergshClient.In, dockergshClient.Out)
		if username == "" {
			username = authConfig.Username
		}
	}

	// 命令行中指定的用户名与配置文件中的不同
	if username != authConfig.Username {
		if password == "" {
			// 保存终端状态
			oldState, _ := terminal.SaveState(dockergshClient.TerminalFd)
			// 输入密码
			fmt.Fprintln(dockergshClient.Out, "Password: ")

			// 禁用终端 回显
			terminal.DisableEcho(dockergshClient.TerminalFd, oldState)

			// 读取密码
			password = readInput(dockergshClient.In, dockergshClient.Out)
			fmt.Fprint(dockergshClient.Out, "\n")

			// 恢复终端原始配置
			_ = terminal.RestoreTerminal(dockergshClient.TerminalFd, oldState)

			if password == "" {
				return fmt.Errorf("error : Password Required")
			}
		}

		if email == "" {
			promptDefault("Email", authConfig.Email)
			email = readInput(dockergshClient.In, dockergshClient.Out)
			if email == "" {
				email = authConfig.Email
			}
		}
	} else {
		password = authConfig.Password
		email = authConfig.Email
	}

	// 重新设置 用户名密码
	authConfig.Username = username
	authConfig.Password = password
	authConfig.Email = email
	authConfig.ServerAddress = serverAddress
	dockergshClient.ConfigFile.Configs[serverAddress] = authConfig

	// 发送 auth 请求
	stream, statueCode, err := dockergshClient.Call("POST", "/auth", dockergshClient.ConfigFile.Configs[serverAddress], false)
	if err != nil {
		return err
	}

	// 如果报错 401 , 从配置文件中删除配置信息
	if statueCode == 401 {
		delete(dockergshClient.ConfigFile.Configs, serverAddress)
		registry.SaveConfig(dockergshClient.ConfigFile)
		return err
	}

	// 将 auth 请求的响应转成 Env
	var out2 engine.Env
	err = out2.Decode(stream)
	// 如果返回值不正确，重新加载 HOME 下的 auth config
	if err != nil {
		dockergshClient.ConfigFile, _ = registry.LoadConfig(os.Getenv("HOME"))
		return err
	}

	// 否则，保存正确的 认证信息 到 auth config 文件
	registry.SaveConfig(dockergshClient.ConfigFile)
	if out2.Get("status") != "" {
		fmt.Fprintf(dockergshClient.Out, "%s\n", out2.Get("Status"))
	}
	return nil
}
