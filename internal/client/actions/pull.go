package actions

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/Nevermore12321/dockergsh/internal/client"
	"github.com/Nevermore12321/dockergsh/internal/registry"
	"github.com/Nevermore12321/dockergsh/pkg/parse"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"net/url"
	"strings"
)

func CmdPullDescription() string {
	return "Pull an image or a repository from the registry"
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

	// 3. 加载配置文件
	dockerClient := client.Client
	err = dockerClient.LoadConfigFile()
	if err != nil {
		log.Warnf("load config file error %v", err)
	}

	// 从配置信息中，匹配当前使用的仓库地址，获取对应的认证信息
	authConfig := dockerClient.ConfigFile.ResolveAuthConfig(hostname)

	pull := func(authConfig registry.AuthConfig) error {
		buff, err := json.Marshal(authConfig)
		if err != nil {
			return err
		}

		registryAuthHeader := []string{
			base64.URLEncoding.EncodeToString(buff),
		}

		// 向仓库发送拉取请求
		return dockerClient.Stream("POST", "/images/create?"+v.Encode(), nil, dockerClient.Out, map[string][]string{
			"X-Registry-Auth": registryAuthHeader,
		})
	}

	if err := pull(authConfig); err != nil {
		if strings.Contains(err.Error(), "Status 401") {
			fmt.Fprintln(dockerClient.Out, "\nPlease login prior to pull:")
			// 先尝试登陆
			if err := CmdLogin(context); err != nil {
				return err
			}
			// 重试
			authConfig = dockerClient.ConfigFile.ResolveAuthConfig(hostname)
			return pull(authConfig)
		}
		return err
	}

	return nil
}
