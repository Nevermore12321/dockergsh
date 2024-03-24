package actions

import (
	"fmt"
	"github.com/Nevermore12321/dockergsh/internal/utils"
	"github.com/urfave/cli/v2"
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
	host := context.String(utils.DockergshHosts)
	protohost := strings.SplitN(host, "://", 2) // 获取通过：//分割的两部分

	// 初始化 dockergshclient
	// 创建Docker Client实例。
	// todo client 使用全局变量
	client := client.NewDockergshClient(context.App.Reader, context.App.Writer, context.App.ErrWriter, protohost[0], protohost[1], tlsConfig)
	fmt.Println(client)
	return nil
}
