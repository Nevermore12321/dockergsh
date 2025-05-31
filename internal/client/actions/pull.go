package actions

import (
	"errors"
	"github.com/Nevermore12321/dockergsh/internal/registry"
	"github.com/Nevermore12321/dockergsh/pkg/parse"
	"github.com/urfave/cli/v2"
	"net/url"
)

var (
	ErrCmdFormat = errors.New("the format of the command you entered is incorrect. Please use -h to check usage")
)

func CmdPullFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:    "pidfile",
			Aliases: []string{"p"},
			Value:   "/var/run/dockergsh.pid",
			Usage:   "Path to use for daemon PID file",
		},
	}
}

// CmdPull 解析参数 docker pull NAME[：TAG]
// 例如 docker pull localhost.localdomain:5000/docker/ubuntu:14.04
func CmdPull(context *cli.Context) error {
	if context.NArg() != 1 {
		return ErrCmdFormat
	}

	var (
		v      = url.Values{}           // Docker Client发送请求给Docker Server时，需要为请求配置URL的查询参数
		remote = context.Args().First() // 镜像的详细 url
	)

	// 设置 Image 完整 url 地址
	v.Set("fromImage", remote)

	// 1. 解析镜像地址仓库地址
	remote, tag := parse.ParseRepositoryTag(remote)

	// 设置 Image tag 地址
	v.Set("tag", tag)

	// 2. 解析镜像仓库中的 hostname + 镜像名称
	hostname, _, err := registry.ResolveRepositoryName(remote)
	if err != nil {
		return err
	}

	return nil
}
