//go:build linux

package cgroups

import (
	log "github.com/sirupsen/logrus"
	"os/exec"
)

var (
	Version string
)

func init() {
	// 检查版本
	if Version == "" {
		_, err := exec.Command("grep", "cgroup2", "/proc/filesystems").CombinedOutput()
		if err != nil { // cgroup v1
			log.Infof("Use Cgroup V1 Version.")
			Version = "CgroupV1"
		} else {
			log.Infof("Use Cgroup V2 Version.")
			Version = "CgroupV2"
		}
	}
}

type Manager interface {
	Apply(pid int) error // 将cgroup配置 apply 到指定 pid 进程
}
