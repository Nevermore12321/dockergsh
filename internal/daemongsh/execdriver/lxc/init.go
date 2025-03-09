package lxc

import "github.com/Nevermore12321/dockergsh/internal/reexec"

func init() {
	reexec.Register("/.dockergshinit", dockerInitializer)
}

func dockerInitializer() {
	initializer()
}

// initializer 是LXC驱动程序的初始函数，在名称空间内运行以设置其他配置
func initializer() {
	// todo
}
