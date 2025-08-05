package configuration

import "io"

type Version string

// 配置表，由 yaml 文件解析得到，配置文件中的字段不能使用 _，因为环境使用 _
type Configuration struct {
	Version       Version       `yaml:"version"`                 // 定义配置文件格式的版本
	Log           Log           `yaml:"log"`                     // 日志
	Storage       Storage       `yaml:"storage"`                 // 镜像存储配置
	Auth          Auth          `yaml:"auth,omitempty"`          // 仓库是否认证信息，可以配置为匿名
	Middleware    Middleware    `yaml:"middleware,omitempty"`    // 配置仓库服务中间件
	HTTP          HTTP          `yaml:"HTTP,omitempty"`          // 配置 http server 相关信息，包括证书，端口等等
	Notifications Notifications `yaml:"notifications,omitempty"` // 通知事件配置
	Redis         Redis         `yaml:"redis,omitempty"`         // 配置 web server 缓存 redis 中间件
	Health        Health        `yaml:"health,omitempty"`        // 健康检查配置
	Catalog       Catalog       `yaml:"catalog,omitempty"`       // Catalog endpoint (/v2/_catalog) 配置，control the maximum number of entries returned by the catalog endpoint
	Proxy         Proxy         `yaml:"proxy,omitempty"`         // 配置代理
	Validation    Validation    `yaml:"validation,omitempty"`    // 配置校验器，例如适合什么平台
	Policy        Policy        `yaml:"policy,omitempty"`        // 配置 policy，白名单机制
}

// Log 表示应用程序内的日志配置。
type Log struct {
}

type Storage struct {
}

type Auth struct {
}

type Middleware struct {
}

type HTTP struct {
}

type Notifications struct {
}

type Redis struct {
}

type Health struct {
}

type Catalog struct {
}

type Proxy struct {
}

type Validation struct {
}

type Policy struct {
}

// todo
func Parse(rd io.Reader) (*Configuration, error) {
	return nil, nil
}
