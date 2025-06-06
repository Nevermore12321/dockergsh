package registry

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// INDEXSERVER 默认仓库地址
const INDEXSERVER = "https://index.docker.io/v1/"

// CONFIGFILE 客户端的配置文件
const CONFIGFILE = ".dockercfg"

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

// LoadConfig 从 rootPath 下加载配置文件 .dockercfg
// rootPath 是当前用户的根路径，例如 /home/gsh
func LoadConfig(rootPath string) (*ConfigFile, error) {
	configFile := &ConfigFile{Configs: make(map[string]AuthConfig), rootPath: rootPath}

	// 判断配置文件是否存在
	cfgFile := filepath.Join(rootPath, CONFIGFILE)
	if _, err := os.Stat(cfgFile); err != nil { // 获取文件失败，返回空的配置信息
		return configFile, err
	}

	// 读取配置文件
	content, err := os.ReadFile(cfgFile)
	if err != nil {
		return configFile, err
	}

	// 解析配置文件
	if err := json.Unmarshal(content, &configFile.Configs); err != nil { // 如果不是结构化的数据，那就按照换行符进行分隔
		// 默认格式
		// 1. =user:pass
		// 2. =email@xx.com
		rowArr := strings.Split(string(content), "\n")
		if len(rowArr) < 2 { // 至少需要配置两个参数，分别是 username/password 与 email
			return configFile, fmt.Errorf("the Auth config file is empty")
		}
		authConfig := AuthConfig{}
		// 第一行按照 = 再分割(用户名 密码）
		origAuth := strings.Split(rowArr[0], "=")
		if len(origAuth) != 2 {
			return configFile, fmt.Errorf("invalid Auth config file")
		}
		// username=password 格式
		authConfig.Username, authConfig.Password, err = decodeAuth(origAuth[1])
		if err != nil {
			return configFile, err
		}

		// 第二行按照 = 再分割（邮箱）
		origEmail := strings.Split(rowArr[1], "=")
		if len(origEmail) != 2 {
			return configFile, fmt.Errorf("invalid Auth config file")
		}
		authConfig.Email = origEmail[1]
		authConfig.ServerAddress = IndexServerAddress()
		configFile.Configs[authConfig.ServerAddress] = authConfig
	} else { // 结构化的数据，直接解析，有可能配置多个仓库的认证信息
		for k, authConfig := range configFile.Configs {
			// auth 表示 user:pass
			authConfig.Username, authConfig.Password, err = decodeAuth(authConfig.Auth)
			if err != nil {
				return configFile, err
			}
			authConfig.Auth = ""
			authConfig.ServerAddress = k
			configFile.Configs[k] = authConfig
		}
	}

	return configFile, nil
}

func decodeAuth(authStr string) (string, string, error) {
	decLen := base64.StdEncoding.DecodedLen(len(authStr)) // 原始数据的长度
	decodedContent := make([]byte, decLen)
	n, err := base64.StdEncoding.Decode(decodedContent, []byte(authStr)) // 解码
	if err != nil {
		return "", "", err
	}

	// 解码后的长度不对
	if n != decLen {
		return "", "", fmt.Errorf("something went wrong decoding auth config")
	}

	// 按照冒号分割user:pass
	authArr := strings.Split(string(decodedContent), ":")
	if len(authArr) != 2 {
		return "", "", fmt.Errorf("invalid auth configuration file")
	}

	// 去掉前后空格
	password := strings.Trim(authArr[1], "\x00")
	return authArr[0], password, nil
}

// ResolveAuthConfig 根据仓库的url解析 hostname
func (configFile *ConfigFile) ResolveAuthConfig(hostName string) AuthConfig {
	// 如果是默认的仓库地址
	if hostName == IndexServerAddress() {
		return configFile.Configs[IndexServerAddress()]
	}

	// 如果之前 LoadConfig 中配置了该仓库的登陆信息
	if c, ok := configFile.Configs[hostName]; ok {
		return c
	}

	// 将 url 解析出 hostname
	convertToHostname := func(url string) string {
		stripped := url
		if strings.HasPrefix(url, "http://") {
			stripped = strings.Replace(url, "http://", "", 1)
		}

		if strings.HasPrefix(url, "https://") {
			stripped = strings.Replace(url, "https://", "", 1)
		}

		nameParts := strings.SplitN(stripped, "/", 2)
		return nameParts[0]
	}

	normalizedHostename := convertToHostname(hostName)
	for registry, config := range configFile.Configs {
		if registryName := convertToHostname(registry); normalizedHostename == registryName {
			return config
		}
	}
	return AuthConfig{}
}
