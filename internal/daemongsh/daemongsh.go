package daemongsh

import "github.com/Nevermore12321/dockergsh/internal/engine"

/*
在 Dockergsh架构中 Daemongsh 支撑着整个后台进程的运行，
同时也统一化管理着 Docker 架构中 graph、graphdriver、execdriver、volumes、Docker 容器等众多资源。
可以说，Dockergsh Daemongsh复杂的运作均由daemongsh对象来调度
*/

type Daemongsh struct {
}

// NewDaemongsh 创建 Daemonize 对象实例
func NewDaemongsh(config *Config, eng *engine.Engine) (*Daemongsh, error) {
	daemongsh, err := NewDaemongshFromDirectory(config, eng)
	if err != nil {
		return nil, err
	}
	return daemongsh, nil
}

// NewDaemongshFromDirectory 具体通过 Config 配置和 engine 对象创建 Daemongsh 对象实例
func NewDaemongshFromDirectory(config *Config, eng *engine.Engine) (*Daemongsh, error) {
	// 1. 应用配置信息
	// 1.1 配置 Dockergsh 容器的 MTU。容器网络接口的最大传输单元（MTU）
	if config.Mtu == 0 { // 表示没有设置
		// 设置一个默认值
		config.Mtu = GetDefaultNetworkMtu()
	}
	// 1.2 检测网桥配置信息
	// 1.3 查验容器间的通信配置
	// 1.4 处理 PID 文件配置

}
