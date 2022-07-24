package v1

import (
	"fmt"
	"github.com/Nevermore12321/dockergsh/cgroup/subsystem"
	"io/ioutil"
	"os"
	"path"
	"strconv"
)

type CpuSubSystem struct {}

func (cs *CpuSubSystem) Name() string {
	return "cpu"
}

func (cs *CpuSubSystem) Set(cgroupPath string, resConf *subsystem.ResourceConfig) error {
	// 获取 cpu subsystem 的路径
	if cpuSubSystemCgroupPath, err := GetCgroupPath(cs.Name(), cgroupPath, true); err != nil {
		return err
	} else {
		// 设置 对应 cgroup 的 cpu 资源限制，也就是修改 cpu.shares 文件
		if resConf.CpuShare != "" {
			if err := ioutil.WriteFile(path.Join(cpuSubSystemCgroupPath, "cpu.shares"), []byte(resConf.CpuShare), 0644); err != nil {
				return fmt.Errorf("set cgroup cpu share fail %v", err)
			}
		}
		return nil
	}
}

// 将一个进程添加到 cgroupPath 对应的 cgroup 中
func (cs *CpuSubSystem) Apply(cgroupPath string, pid int) error {
	if cpuSubSystemCgroupPath, err := GetCgroupPath(cs.Name(), cgroupPath, false); err != nil {
		return fmt.Errorf("get cgroup %s error: %v", cgroupPath, err)
	} else {
		// 将进程号 pid 写入到 cgroup 的虚拟文件系统对应目录下的 tasks 文件中
		if err := ioutil.WriteFile(path.Join(cpuSubSystemCgroupPath, "tasks"), []byte(strconv.Itoa(pid)), 0644); err != nil {
			return fmt.Errorf("set cgroup proc fail %v", err)
		} else {
			return nil
		}
	}
}

// 删除 cgroupPath 对应的 cgroup
func (cs *CpuSubSystem) Remove(cgroupPath string) error {
	if cpuSubSystemCgroupPath, err := GetCgroupPath(cs.Name(), cgroupPath, false); err != nil {
		return err
	} else {
		// 产出 cgroup 的目录，便是删除了 cgroup
		return os.RemoveAll(cpuSubSystemCgroupPath)
	}
}
