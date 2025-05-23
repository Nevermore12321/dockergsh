//go:build linux

package v2

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"os"
	"path"
	"strings"
)

var (
	CgroupV2RootPathPrefix = "/sys/fs/cgroup"
	Root                   string
)

// FindCgroupMountPoint FindCgroupMountPoint("memory"),这里返回具体某个 cgroup 挂载的根路径
func FindCgroupMountPoint() (string, error) {
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

	return CgroupV2RootPathPrefix + userCgroupPath + "/dockergsh", nil
}

// GetCgroupPath 获取 cgroup 的绝对路径
// 注意 cgroup v2 版本，已经把所有的 hierarchy 都统一到 根下，因此只有一个 hierarchy。
// 因此 所有的 subsystem 都有统一的路径，与 v1 不同
func GetCgroupPath(subsystem string, autoCreate bool) (string, error) {
	var cgroupRoot string
	var err error

	if Root == "" {
		Root, err = FindCgroupMountPoint()
		if err != nil {
			return "", fmt.Errorf("find cgroup mount point error %v", err)
		}

		cgroupRoot = CgroupV2RootPathPrefix
	} else {
		cgroupRoot = Root
	}
	log.Infof("Cgroup Root: %s", cgroupRoot)

	// os.Stat返回描述文件 f 的 FileInfo 类型值。如果出错，错误底层类型是 *PathError
	_, err = os.Stat(cgroupRoot)

	// 如果目录不存在就创建
	if err == nil || (autoCreate && os.IsNotExist(err)) {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(cgroupRoot, 0755); err != nil {
				return "", fmt.Errorf("error create cgroup %v", err)
			}
			// 这个文件内容应是cgroup.controllers的子集。其作用是限制在当前cgroup目录层级下创建的子目录中的cgroup.controllers内容。
			//就是说，子层级的cgroup资源限制范围被上一级的cgroup.subtree_control文件内容所限制。
			if err := os.WriteFile(
				path.Join(cgroupRoot, "cgroup.subtree_control"),
				[]byte("+cpu +cpuset +memory +io +pids"),
				0644); err != nil {
				return "", fmt.Errorf("set cgroup subtree_control fail %v", err)
			}
		}
		if subsystem == "" {
			return cgroupRoot, nil
		}
		return path.Join(cgroupRoot, subsystem), nil
	} else {
		return "", fmt.Errorf("cgroup path error %v", err)
	}
}
