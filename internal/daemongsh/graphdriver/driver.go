package graphdriver

import (
	"errors"
	"github.com/Nevermore12321/dockergsh/internal/utils"
	"os"
	"path"
)

/*
加载并配置存储驱动 graphdriver
目的在于：使得 Daemongsh 创建 Dockergsh 镜像管理所需的驱动环境。
graphdriver 用于完成 Dockergsh 镜像的管理，包括获取、存储以及容器 rootfs 的构建等
*/

// InitFunc 各个不同的 存储 driver ，都需要实现初始化函数
type InitFunc func(root string, options []string) (Driver, error)

var (
	DefaultDriver string                                             // 默认的 graphdriver
	priority      = []string{"aufs", "btrfs", "devicemapper", "vfs"} // driver 优先级列表
	drivers       map[string]InitFunc                                // 所有 graphdriver 的 名称-InitFunc 的映射关系
)

var (
	ErrNotSupport     = errors.New("driver not supported")                                       // 此驱动不支持
	ErrPrerequisite   = errors.New("prerequisites for driver not satisfied (wrong filesystem?)") // 先决条件不满足
	ErrIncompatibleFs = errors.New("file system is unsupported for this graph driver")           // 文件系统不兼容此驱动
	ErrNoGraphDrivers = errors.New("no graph drivers that meet the conditions")
)

type Driver interface {
	String() string                 //  driver 的打印输出格式
	Create(id, parent string) error // 创建存储层
	Remove(id string) error         // 删除存储层
	Exists(id string) bool          // 判断存储层是否已经存在

	// todo put get exists status cleanups
}

func init() {
	drivers = make(map[string]InitFunc)
}

// GetDriver 通过 driver 的名称、根路径，获取 graph-driver
func GetDriver(name, home string, options []string) (Driver, error) {
	// 从 drivers 映射表中找到与 name 名称对应的 initFunc 初始化函数
	if initFunc, ok := drivers[name]; ok {
		// 如果存在，直接初始化创建一个 driver，根路径为  root/name
		return initFunc(path.Join(home, name), options)
	}
	// 如果不存在，返回错误
	return nil, ErrNotSupport
}

// New 加载graph的存储驱动
func New(root string, options []string) (Driver, error) {
	var (
		driver Driver
		err    error
	)
	// 1. 遍历数组选择 graphdriver

	// 1.1 DOCKERGSH_GRAPHDRIVER 设置 driver
	// 1.2 DefaulDriver
	graphDriver := os.Getenv(utils.GraphDriver)

	// 优先使用用户自定义的 driver 也就是环境变量的 DOCKERGSH_GRAPHDRIVER
	for _, name := range []string{graphDriver, DefaultDriver} {
		if name != "" {
			// 如果已经配置，直接初始化并返回 driver
			return GetDriver(root, name, options)
		}
	}

	// 2. 遍历优先级数组priority选择graphdriver
	// 2.1 优先级数组依次是： aufs、btrfs、devicemapper、vfs(优先级从高到低)
	// 在没有指定以及默认的驱动时，从优先级数组中选择驱动，目前优先级最高的为aufs
	for _, name := range priority { // 找到能满足的，优先级最高的 driver
		driver, err = GetDriver(name, root, options)
		if err != nil {
			if err == ErrNotSupport || err == ErrPrerequisite || err == ErrIncompatibleFs { // 如果是不兼容，继续找下一个 driver
				continue
			}
			return nil, err
		}
		return driver, nil
	}

	// 3. 从已经注册的drivers数组中选择graphdriver
	// 在aufs、btrfs、devicemapper和vfs四个不同类型驱动的init函数中，它们均向graphdriver的drivers数组注册了相应的初始化方法
	// 在没有优先级数组的时候，同样可以通过注册的驱动来选择具体的graphdriver
	for _, initFunc := range drivers {
		if driver, err = initFunc(root, options); err != nil {
			if err == ErrNotSupport || err == ErrPrerequisite || err == ErrIncompatibleFs { // 如果是不兼容，继续找下一个 driver
				continue
			}
			return nil, err
		}
		return driver, nil
	}

	// 4. 如果都没找到，返回错误
	return nil, ErrNoGraphDrivers
}
