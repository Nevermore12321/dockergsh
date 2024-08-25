package lxc

import "fmt"

const DriverName = "lxc"

type Driver struct {
	root       string // dockergsh 的根目录，默认 /var/lib/dockergsh
	initPath   string // dockergshInit 初始化目录 /var/lib/docker/init/dockerinit-[VERSION]
	apparmor   bool   // 是否开启了 AppArmor 权限控制功能
	sharedRoot bool   // 是否共享 root 目录
}

// NewDriver todo
func NewDriver(root, initPath string, appArmor bool) (*Driver, error) {
	return nil, nil
}

// Name todo
func (d *Driver) Name() string {
	//version := d.version()
	version := "0.1"
	return fmt.Sprintf("%s-%s", DriverName, version)
}
