package kernel

// 获取到运行 dockergsh 所在服务器的 kernel 版本
type KernelVersionInfo struct {
	Kernel int    // kernel 版本号
	Major  int    // kernel 的 major 版本号
	Minor  int    // kernel 的 minor 版本号
	Flavor string //
}
