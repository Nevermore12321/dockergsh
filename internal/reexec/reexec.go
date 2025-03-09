package reexec

import (
	"fmt"
	"os"
)

// dockerinit：是在新 namespace 中第一个运行的内容，作用是设置新 namespace 中的挂载资源，初始化容器内的网络栈等。
// 完成的属于容器层系统环境的初始化工作。

// 以网络 namespace 为例：
// 当 DockerDaemon 创建容器时，仅仅是 fork 了一个进程，那么如何让这个进程的 net namespace 中包含一个可用的网络栈（虚拟网络设备 veth）呢？
// 为容器内部初始化网络栈的角色就是 dockerinit
// dockerinit 会获取 Docker Daemon 传递来的网络信息，并用来初始化这个容器的 net namespace。保证后续的进程拥有足够的网络能力。

// 注册的初始化函数
var registeredInitializers = make(map[string]func())

// Init docker run 和 exec 执行的第一步骤
func Init() bool {
	// 从已经注册的初始化函数表中寻找
	initializer, exists := registeredInitializers[os.Args[0]]
	if exists {
		initializer()
		return true
	}
	return false
}

func Register(name string, initializer func()) {
	if _, exists := registeredInitializers[name]; exists {
		panic(fmt.Sprintf("reexec func already registred under name %q", name))
	}
	registeredInitializers[name] = initializer
}
