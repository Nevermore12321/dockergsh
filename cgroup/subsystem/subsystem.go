package subsystem


// 用于传递资源限制配置的结构体，包含内存限制，CPU时间片去重，CPU核心数
type ResourceConfig struct {
	MemoryLimit string
	CpuShare	string
	CpuSet		string
}

// Subsystem 接口，每个 subsystem 需要实现四个接口
// 将 cgroup 抽象成 path 路径，原因就是 cgroup 在 hierarchy 的路径，其实就是虚拟文件路径
type Subsystem interface {
	// Subsystem 的名字，例如 cpu、memory
	Name() string
	// 设置某个 cgroup 在这个 subsystem 中的资源限制，也就是修改具体的配置文件，v1 与 v2 略有不同
	Set(cgroupPath string, resConf *ResourceConfig) error
	// 将进程添加到某个 cgroup 中
	Apply(cgroupPath string, pid int) error
	// 删除某个 cgroup
	Remove(cgroupPath string) error
}