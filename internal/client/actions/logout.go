package actions

import (
	"fmt"
	"github.com/Nevermore12321/dockergsh/internal/client"
	"github.com/Nevermore12321/dockergsh/internal/registry"
	"github.com/Nevermore12321/dockergsh/internal/utils"
	"github.com/urfave/cli/v2"
)

func CmdLogoutDescription() string {
	return fmt.Sprintf("Usage: docker [OPTIONS] COMMAND [arg...]\n -H=[unix://%s]: tcp://host:port to bind/connect to or unix://path/to/socket to use\n\nA self-sufficient runtime for linux containers.\n\nCommands:\n", utils.DefaultUnixSocket)
}

// CmdLogin "Docker登录"：登出
func CmdLogout(context *cli.Context) error {
	// 默认的登出仓库地址
	serverAddress := registry.IndexServerAddress()
	if context.Args().Len() > 0 {
		serverAddress = context.Args().Get(0)
	}

	dockergshClient := client.Client
	_ = dockergshClient.LoadConfigFile()

	// 如果配置文件中没有当前仓库地址的认证信息
	if _, ok := dockergshClient.ConfigFile.Configs[serverAddress]; !ok {
		fmt.Fprintf(dockergshClient.Out, "Not logged in to %s\n", serverAddress)
	} else { // 否则
		fmt.Fprintf(dockergshClient.Out, "Remove login credentials for %s\n", serverAddress)
		// 删除相关的认证信息
		delete(dockergshClient.ConfigFile.Configs, serverAddress)

		// 保存配置文件
		if err := registry.SaveConfig(dockergshClient.ConfigFile); err != nil {
			return fmt.Errorf("failed to save docker config: %v", err)
		}
	}
	return nil
}
