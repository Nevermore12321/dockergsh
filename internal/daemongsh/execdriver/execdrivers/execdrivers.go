package execdrivers

import (
	"fmt"
	"github.com/Nevermore12321/dockergsh/internal/daemongsh/execdriver"
	"github.com/Nevermore12321/dockergsh/internal/daemongsh/execdriver/lxc"
	"github.com/Nevermore12321/dockergsh/internal/daemongsh/execdriver/native"
	"github.com/Nevermore12321/dockergsh/pkg/sysinfo"
)

// NewDriver 根据 execdriver 名称，创建 execdirver 实例
// - config.ExecDriver：Docker运行时中用户指定使用的execdriver类型，在默认配置文件中值为native。用户也可以在启动DockerDaemon将这个值配置为lxc，则导致Docker使用lxc类型的驱动执行Docker容器的内部操作。
// - config.Root：Docker运行时的root路径，默认配置文件中为/var/lib/docker。
// - sysInitPath：系统中存放dockerinit二进制文件的路径，一般为/var/lib/docker/init/dockerinit-[VERSION]。
// - sysInfo：系统功能信息，包括：容器的内存限制功能，交换区内存限制功能，数据转发功能，以及AppArmor安全功能等
func NewDriver(name, root, initPath string, sysInfo *sysinfo.SysInfo) (execdriver.Driver, error) {
	switch name {
	case "lxc":
		return lxc.NewDriver(root, initPath, sysInfo.AppArmor)
	case "native":
		return native.NewDriver(root, initPath)
	}
	return nil, fmt.Errorf("unknown exec driver %s", name)
}
