package registry

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// 在 engine 对象对外暴露的API信息中添加 dockergsh registry 的信息。
// 若 registry.NewService（）被成功安装，则会有两个相应的处理方法注册至 engine:
// 1. Dockergsh Daemon 通过 Docker Client 提供的认证信息向 registry 发起认证请求；
// 2. search，在公有registry上搜索指定的镜像，目前公有的registry只支持 Docker Hub。

var (
	ErrInvalidRepositoryName = errors.New("Invalid repository name (ex: \"registry.domain.tld/myrepos\")")
)

// ResolveRepositoryName 从镜像url中解析出 仓库地址 + 镜像名
// 例如： localhost.localdomain:5000/docker/ubuntu 解析结果 localhost.localdomain:5000 + docker/ubuntu
func ResolveRepositoryName(reposName string) (string, string, error) {
	// 镜像地址不能带 http://
	if strings.Contains(reposName, "://") {
		return "", "", ErrInvalidRepositoryName
	}

	// 根据 / 解析 host 和 镜像名，分成两部分，第一个 / 之前的就是 host，后面就是镜像名
	nameParts := strings.SplitN(reposName, "/", 2)
	// samalba/hipache or ubuntu，不带 host 仓库地址
	if len(nameParts) == 1 || (!strings.Contains(nameParts[0], ".") && !strings.Contains(nameParts[0], ":") && nameParts[0] != "localhost") {
		err := validateRepositoryName(reposName)
		// 没有仓库地址，默认使用 公共的镜像地址
		return IndexServerAddress(), reposName, err
	}

	// 否则，就是镜像 url 中既携带了 仓库地址，也带了 repostoryName
	hostName := nameParts[0]
	reposName = nameParts[1]

	// 默认使用 index.docker.io，不需要带
	if strings.Contains(hostName, "index.docker.io") {
		return "", "", fmt.Errorf("invalid repository name, try \"%s\" instead", reposName)
	}

	// 校验 repositoryName
	if err := validateRepositoryName(reposName); err != nil {
		return "", "", err
	}
	return hostName, reposName, nil

}

// 检查镜像的 RepositoryName 是否正确
// 例如 samalba/hipache ，RepositoryName 为 samalba
// 例如 ubuntu， RepositoryName 为 library
func validateRepositoryName(repositoryName string) error {
	var (
		name      string
		namespace string
	)

	nameParts := strings.SplitN(repositoryName, "/", 2)
	if len(nameParts) < 2 { // 直接使用 samalba/hipache or ubuntu
		namespace = "library"
		name = nameParts[0]
	} else {
		namespace = nameParts[0]
		name = nameParts[1]
	}

	// 校验 RepositoryName
	validNamespace := regexp.MustCompile(`^([a-z0-9_]{4,30})$`)
	if !validNamespace.MatchString(namespace) {
		return fmt.Errorf("invalid namespace name (%s), only [a-z0-9_] are allowed, size between 4 and 30", namespace)
	}

	// 校验 ImageName
	validName := regexp.MustCompile(`^([a-z0-9_]+)$`)
	if !validName.MatchString(name) {
		return fmt.Errorf("Invalid repository name (%s), only [a-z0-9-_.] are allowed", name)
	}
	return nil

}
