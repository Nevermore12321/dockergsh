package cgroup

import (
	"github.com/Nevermore12321/dockergsh/cgroup/subsystem"
	subSysV1 "github.com/Nevermore12321/dockergsh/cgroup/subsystem/v1"
	subSysV2 "github.com/Nevermore12321/dockergsh/cgroup/subsystem/v2"
	log "github.com/sirupsen/logrus"
)

/*
把所有不同的 subsystem 中的 cgroup 管理起来，并与容器建立关系
 */
type CgroupManager struct {
	Path 		string				//  cgroup在hierarchy中的路径 相当于创建的cgroup目录相对于root cgroup目录的路径
	Resource 	*subsystem.ResourceConfig	// 资源配置
}

//  工厂函数
func NewCgroupManager(path string) *CgroupManager {
	return &CgroupManager{
		Path: path,
	}
}

// 将进程加入到 cgroup v1 的每个 cgroup 中
func (cm *CgroupManager) ApplyV1(pid int) error {
	for _, subSystemIns := range subSysV1.SubsystemIns {
		if err := subSystemIns.Apply(cm.Path, pid); err != nil {
			return err
		}
	}
	return nil
}

// 设置各个挂载 subsystem 的 cgroup 的资源限制
func (cm *CgroupManager) SetV1(resConf *subsystem.ResourceConfig) error {
	for _, subSystemIns := range subSysV1.SubsystemIns {
		if err := subSystemIns.Set(cm.Path, resConf); err != nil {
			return err
		}
	}
	return nil
}

// 释放各个 subsystem 挂在中的 cgroup
func (cm *CgroupManager) DestroyV1() error {
	for _, subsystemIns := range subSysV1.SubsystemIns {
		if err := subsystemIns.Remove(cm.Path); err != nil {
			log.Warnf("remove cgroup fail %v", err)
			return err
		}
	}
	return nil
}


// 将进程加入到 cgroup v2 的每个 cgroup 中
func (cm *CgroupManager) ApplyV2(pid int) error {
	for _, subSystemIns := range subSysV2.SubsystemIns {
		if err := subSystemIns.Apply(cm.Path, pid); err != nil {
			return err
		}
	}
	return nil
}

// 设置各个挂载 subsystem 的 cgroup 的资源限制
func (cm *CgroupManager) SetV2(resConf *subsystem.ResourceConfig) error {
	for _, subSystemIns := range subSysV2.SubsystemIns {
		if err := subSystemIns.Set(cm.Path, resConf); err != nil {
			return err
		}
	}
	return nil
}

// 释放各个 subsystem 挂在中的 cgroup
func (cm *CgroupManager) DestroyV2() error {
	for _, subsystemIns := range subSysV2.SubsystemIns {
		if err := subsystemIns.Remove(cm.Path); err != nil {
			log.Warnf("remove cgroup fail %v", err)
			return err
		}
	}
	return nil
}