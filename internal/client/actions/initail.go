package actions

import (
	"github.com/Nevermore12321/dockergsh/internal/utils"
	"github.com/urfave/cli/v2"
	"os"
	"strings"
)
import "github.com/Nevermore12321/dockergsh/internal/client"

func CmdClientInitial(context *cli.Context) error {
	// docker client

	// 初始化 Tls 配置
	tlsConfig, err := client.PreCheckConfTLS(context)
	if err != nil {
		return err
	}

	// 从环境变量中获取 host 信息
	host := os.Getenv(utils.DockergshHosts)
	protohost := strings.SplitN(host, "://", 2) // 获取通过：//分割的两部分

	// 初始化 dockergshclient
	// 创建Docker Client实例。
	// client 使用全局变
	client.Client = client.NewDockergshClient(context.App.Reader, context.App.Writer, context.App.ErrWriter, protohost[0], protohost[1], tlsConfig)
	return nil
}
