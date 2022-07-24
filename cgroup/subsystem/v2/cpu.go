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

type CpuSubSystem struct{}

func (cs *CpuSubSystem) Name() string {
	return "cpu"
}

func (cs *CpuSubSystem) Set(cgroupPath string, resConf *subsystem.ResourceConfig) error {
	if cpuSubSystemCgroupPath, err := GetCgroupPath(cgroupPath, true); err != nil {
		return err
	} else {
		// v2 的cpu配置在文件 cpu.max, 格式为 $MAX $PERIOD，表示在一段 PERIOD 时间内，占用了多少
		if resConf.CpuShare != "" {
			per, _ := strconv.Atoi(resConf.CpuShare)
			// 这里统一规划为 0-100，表示占用 %多少
			if per < 0 || per > 100 {
				return fmt.Errorf("set cgroup cpu fail")
			}
			cpuShare := fmt.Sprintf("%d %d", per*1000, 100000)
			if err := ioutil.WriteFile(
				path.Join(cpuSubSystemCgroupPath, "cpu.max"),
				[]byte(cpuShare),
				0644); err != nil {
				return fmt.Errorf("set cgroup cpu fail %v", err)
			}
		}
		return nil
	}
}

func (cs *CpuSubSystem) Apply(cgroupPath string, pid int) error {
	if cpuSubSystemCgroupPath, err := GetCgroupPath(cgroupPath, true); err != nil {
		return err
	} else {
		log.Info(cgroupPath, cpuSubSystemCgroupPath, pid)
		if err := ioutil.WriteFile(
			path.Join(cpuSubSystemCgroupPath, "cgroup.procs"),
			[]byte(strconv.Itoa(pid)), 0644); err != nil {
			return fmt.Errorf("set cgroup proc fail %v", err)
		}
		return nil
	}
}

func (cs *CpuSubSystem) Remove(cgroupPath string) error {
	if cpuSubSystemCgroupPath, err := GetCgroupPath(cgroupPath, true); err != nil {
		return err
	} else {
		return os.RemoveAll(cpuSubSystemCgroupPath)
	}
}
