package sysinfo

import (
	"github.com/Nevermore12321/dockergsh/external/libcontainer/cgroups"
	"os"

	log "github.com/sirupsen/logrus"
	"path/filepath"
)

const (
	defaultApparmorPath = "/sys/kernel/security/apparmor"
)

// SysInfo 获取系统功能信息
type SysInfo struct {
	MemoryLimit            bool // 容器的内存限制功能
	SwapLimit              bool // 交换区内存限制功能
	IPv4ForwardingDisabled bool // 数据转发功能
	AppArmor               bool // AppArmor安全功能
}

func New(quiet bool) *SysInfo {
	sysInfo := &SysInfo{}
	if cgroupMemoryMountPoint, err := cgroups.FindCgroupMountPoint("memory"); err != nil {
		if !quiet {
			log.Warnf("WARNING: Failed to find memory mount point: %s", err)
		}
	} else {
		// check MemoryLimit and SwapLimit
		var err1, err2, err3 error
		switch cgroups.Version {
		case "CgroupV1":
			_, err1 = os.ReadFile(filepath.Join(cgroupMemoryMountPoint, "memory.limit_in_bytes"))
			_, err2 = os.ReadFile(filepath.Join(cgroupMemoryMountPoint, "memory.soft_limit_in_bytes"))
			_, err3 = os.ReadFile(filepath.Join(cgroupMemoryMountPoint, "memory.memsw.limit_in_bytes"))
		case "CgroupV2":
			_, err1 = os.ReadFile(filepath.Join(cgroupMemoryMountPoint, "memory.max"))
			_, err2 = os.ReadFile(filepath.Join(cgroupMemoryMountPoint, "memory.low"))
			_, err3 = os.ReadFile(filepath.Join(cgroupMemoryMountPoint, "memory.swap.max"))
		}
		sysInfo.MemoryLimit = err1 == nil && err2 == nil
		if !sysInfo.MemoryLimit && !quiet {
			log.Infof("WARNING: Your kernel does not support cgroup memory limit.")
		}

		sysInfo.SwapLimit = err3 == nil
		if !sysInfo.SwapLimit && !quiet {
			log.Infof("WARNING: Your kernel does not support cgroup swap limit.")
		}
	}

	// Check AppArmor
	// 目前有两个子系统是专门为Linux系统的安全而设计的，分别是：
	// - Security-Enhanced Linux（SELinux）
	// - AppArmor
	// 都提供了应用程序之间的隔离，从而限制了黑客可以用来访问系统的攻击平面
	// 判断是否有 apparmor 的安装目录
	if _, err := os.Stat(defaultApparmorPath); os.IsNotExist(err) {
		sysInfo.AppArmor = false
	} else {
		sysInfo.AppArmor = true
	}

	return sysInfo
}
