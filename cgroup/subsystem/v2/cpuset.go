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

type CpuSetSubSystem struct{}

func (css *CpuSetSubSystem) Name() string {
	return "cpuset"
}

func (css *CpuSetSubSystem) Set(cgroupPath string, resConf *subsystem.ResourceConfig) error {
	if cpuSetSubSystemCgroupPath, err := GetCgroupPath(cgroupPath, true); err != nil {
		return err
	} else {
		// v2 的cpu配置在文件 cpu.max, 格式为 $MAX $PERIOD，表示在一段 PERIOD 时间内，占用了多少
		if resConf.CpuSet != "" {
			if err := ioutil.WriteFile(
				path.Join(cpuSetSubSystemCgroupPath, "cpuset.cpus"),
				[]byte(resConf.CpuSet),
				0644); err != nil {
				return fmt.Errorf("set cgroup cpuset fail %v", err)
			}
		}
		return nil
	}
}

func (css *CpuSetSubSystem) Apply(cgroupPath string, pid int) error {
	if cpuSetSubSystemCgroupPath, err := GetCgroupPath(cgroupPath, true); err != nil {
		return err
	} else {
		log.Info(cgroupPath, cpuSetSubSystemCgroupPath, pid)
		if err := ioutil.WriteFile(
			path.Join(cpuSetSubSystemCgroupPath, "cgroup.procs"),
			[]byte(strconv.Itoa(pid)), 0644); err != nil {
			return fmt.Errorf("set cgroup proc fail %v", err)
		}
		return nil
	}
}

func (css *CpuSetSubSystem) Remove(cgroupPath string) error {
	if cpuSetSubSystemCgroupPath, err := GetCgroupPath(cgroupPath, true); err != nil {
		return err
	} else {
		return os.RemoveAll(cpuSetSubSystemCgroupPath)
	}
}
