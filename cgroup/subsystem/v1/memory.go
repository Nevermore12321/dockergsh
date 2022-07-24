package v1

import (
	"fmt"
	"github.com/Nevermore12321/dockergsh/cgroup/subsystem"
	"io/ioutil"
	"os"
	"path"
	"strconv"
)

type MemorySubSystem struct {}

func (ms *MemorySubSystem) Name() string {
	return "memory"
}

func (ms *MemorySubSystem) Set(cgroupPath string, resConf *subsystem.ResourceConfig) error {
	// 获取 cpuset subsystem 的路径
	if memorySubSystemCgroupPath, err := GetCgroupPath(ms.Name(), cgroupPath, true); err != nil {
		return err
	} else {
		// 设置 对应 cgroup 的 cpu 资源限制，也就是修改 memory.limit_in_bytes 文件
		if resConf.CpuShare != "" {
			if err := ioutil.WriteFile(path.Join(memorySubSystemCgroupPath, "memory.limit_in_bytes"), []byte(resConf.MemoryLimit), 0644); err != nil {
				return fmt.Errorf("set cgroup memory fail %v", err)
			}
		}
		return nil
	}
}

// 将一个进程添加到 cgroupPath 对应的 cgroup 中
func (ms *MemorySubSystem) Apply(cgroupPath string, pid int) error {
	if memorySubSystemCgroupPath, err := GetCgroupPath(ms.Name(), cgroupPath, false); err != nil {
		return fmt.Errorf("get cgroup %s error: %v", cgroupPath, err)
	} else {
		// 将进程号 pid 写入到 cgroup 的虚拟文件系统对应目录下的 tasks 文件中
		if err := ioutil.WriteFile(path.Join(memorySubSystemCgroupPath, "tasks"), []byte(strconv.Itoa(pid)), 0644); err != nil {
			return fmt.Errorf("set cgroup proc fail %v", err)
		} else {
			return nil
		}
	}
}

// 删除 cgroupPath 对应的 cgroup
func (ms *MemorySubSystem) Remove(cgroupPath string) error {
	if memorySubSystemCgroupPath, err := GetCgroupPath(ms.Name(), cgroupPath, false); err != nil {
		return err
	} else {
		// 产出 cgroup 的目录，便是删除了 cgroup
		return os.RemoveAll(memorySubSystemCgroupPath)
	}
}
