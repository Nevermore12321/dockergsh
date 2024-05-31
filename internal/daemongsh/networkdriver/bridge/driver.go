package bridge

import "github.com/Nevermore12321/dockergsh/internal/engine"

// InitDriver builtins 注册的 init_networkdriver job 的handler处理函数
// 1. 获取为Docker服务的网络设备地址。
// 2. 创建指定IP地址的网桥。
// 3. 配置网络iptables规则。
// 4. 另外还为 eng 对象注册了多个 Handler，如 allocate_interface、release_interface、allocate_port以及link等。
func InitDriver(job *engine.Job) engine.Status {
	// todo
	return 0
}
