package registry

const INDEXSERVER = "https://index.docker.io/v1/"

// AuthConfig 与 registry 连接的校验信息
type AuthConfig struct {
	Username      string `json:"username,omitempty"`
	Password      string `json:"password,omitempty"`
	Auth          string `json:"auth,omitempty"`
	Email         string `json:"email,omitempty"`
	ServerAddress string `json:"serverAddress,omitempty"`
}

// ConfigFile 最终会落盘，存储仓库配置信息
type ConfigFile struct {
	Configs  map[string]AuthConfig `json:"configs,omitempty"` // 每一个仓库，对应一组校验信息
	rootPath string
}

func IndexServerAddress() string {
	return INDEXSERVER
}
