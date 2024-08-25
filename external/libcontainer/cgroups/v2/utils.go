//go:build linux

package v2

import (
	"io"
	"os"
	"strings"
)

const (
	CgroupV2RootPathPrefix = "/sys/fs/cgroup/dockergsh"
)

// FindCgroupMountPoint FindCgroupMountPoint("memory"),这里返回具体某个 cgroup 挂载的根路径
func FindCgroupMountPoint(name string) (string, error) {
	// 获取 /proc/self/cgroup 当前用户的 cgroup 相对路径
	file, err := os.Open("/proc/self/cgroup")
	if err != nil {
		return "", err
	}
	defer file.Close()

	text, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}

	// 获取到 /user.slice/user-0.slice/session-11.scope ，将最后的 session 路径去掉
	splitFields := strings.Split(string(text), ":")
	pathArr := strings.Split(splitFields[len(splitFields)-1], "/")
	userCgroupPath := strings.Join(pathArr[0:len(pathArr)-1], "/")

	return CgroupV2RootPathPrefix + userCgroupPath + name, nil
}
