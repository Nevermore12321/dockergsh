package v2

import (
	"fmt"
	"github.com/Nevermore12321/dockergsh/cgroup/subsystem"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"path"
	"strconv"
)

type MemorySubSystem struct{}

func (ms *MemorySubSystem) Name() string {
	return "memory"
}

func (ms *MemorySubSystem) Set(cgroupPath string, resConf *subsystem.ResourceConfig) error {
	if memorySubSystemCgroupPath, err := GetCgroupPath(cgroupPath, true); err != nil {
		return err
	} else {
		// v2 的cpu配置在文件 cpu.max, 格式为 $MAX $PERIOD，表示在一段 PERIOD 时间内，占用了多少
		if resConf.MemoryLimit != "" {
			if err := ioutil.WriteFile(
				path.Join(memorySubSystemCgroupPath, "memory.max"),
				[]byte(resConf.MemoryLimit),
				0644); err != nil {
				return fmt.Errorf("set cgroup memory fail %v", err)
			}
		}
		return nil
	}
}

func (ms *MemorySubSystem) Apply(cgroupPath string, pid int) error {
	if memorySubSystemCgroupPath, err := GetCgroupPath(cgroupPath, true); err != nil {
		return err
	} else {
		log.Info(cgroupPath, memorySubSystemCgroupPath, pid)
		if err := ioutil.WriteFile(
			path.Join(memorySubSystemCgroupPath, "cgroup.procs"),
			[]byte(strconv.Itoa(pid)), 0644); err != nil {
			return fmt.Errorf("set cgroup proc fail %v", err)
		}
		return nil
	}
}

func (ms *MemorySubSystem) Remove(cgroupPath string) error {
	if memorySubSystemCgroupPath, err := GetCgroupPath(cgroupPath, true); err != nil {
		return err
	} else {
		return os.RemoveAll(memorySubSystemCgroupPath)
	}
}
