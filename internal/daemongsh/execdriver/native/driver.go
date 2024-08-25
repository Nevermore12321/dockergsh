package native

import "fmt"

const (
	DriverName = "native"
	Version    = "0.2"
)

// todo
type Driver struct {
	root     string // dockergsh 的根目录，默认 /var/lib/dockergsh
	initPath string // dockergshInit 初始化目录 /var/lib/docker/init/dockerinit-[VERSION]
}

func NewDriver(root string, initPath string) (*Driver, error) {
	return nil, nil
}

func (d *Driver) Name() string {
	return fmt.Sprintf("%s-%s", DriverName, Version)
}
