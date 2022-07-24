package v2

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

var (
	CgroupV2RootPathPrefix string = "/sys/fs/cgroup"
	Root string
)

// 获取 cgroup 的绝对路径
// 注意 cgroup v2 版本，已经把所有的 hierarchy 都统一到 根下，因此只有一个 hierarchy。
// 因此 所有的 subsystem 都有统一的路径，与 v1 不同
func GetCgroupPath(cgroupPath string, autoCreate bool) (string, error) {
	var cgroupRoot string

	if cgroupPath != "" {
		Root = FindCgroupMountpoint()
		cgroupRoot = CgroupV2RootPathPrefix
	} else {
		cgroupRoot = Root
		log.Infof("Cgroup Root: %s", cgroupPath)
	}

	// os.Stat返回描述文件 f 的 FileInfo 类型值。如果出错，错误底层类型是 *PathError
	_, err := os.Stat(path.Join(cgroupRoot, cgroupPath))

	// 如果目录不存在就创建
	if err == nil || (autoCreate && os.IsNotExist(err)) {
		if os.IsNotExist(err) {
			if err := os.Mkdir(path.Join(cgroupRoot, cgroupPath), 0755); err != nil {
				return "", fmt.Errorf("error create cgroup %v", err)
			}
		}
		return path.Join(cgroupRoot, cgroupPath), nil
	} else {
		return "", fmt.Errorf("cgroup path error %v", err)
	}
}


// FindCgroupMountPoint("memory"),这里返回具体某个 cgroup 挂载的根路径
func FindCgroupMountpoint() string {
	// 获取 /proc/self/cgroup 当前用户的 cgroup 相对路径
	file, err := os.Open("/proc/self/cgroup")
	if err != nil {
		return ""
	}

	defer file.Close()

	text, err := ioutil.ReadAll(file)
	if err != nil {
		return ""
	}

	// 获取到 /user.slice/user-0.slice/session-11.scope ，将最后的 session 路径去掉
	split_fields := strings.Split(string(text), ":")
	pathArr := strings.Split(split_fields[len(split_fields) - 1], "/")
	userCgroupPath := strings.Join(pathArr[0:len(pathArr)-1], "/")

	return CgroupV2RootPathPrefix + userCgroupPath
}

