package actions

import (
	"fmt"
	"github.com/Nevermore12321/dockergsh/external/registry/libs/configuration"
	"os"
)

func resolveConfiguration(filePath string) (*configuration.Configuration, error) {
	// 获取 配置文件 路径
	var configurationPath string
	if filePath != "" {
		configurationPath = filePath
	} else {
		configurationPath = os.Getenv("REGISTRY_CONFIGURATION_PATH")
	}

	if configurationPath == "" {
		return nil, fmt.Errorf("configuration path unspecified")
	}

	// 读取配置文件
	fp, err := os.Open(configurationPath)
	if err != nil {
		return nil, err
	}
	defer fp.Close()

	// 解析配置文件到对象 Configuration
	config, err := configuration.Parse(fp)
	if err != nil {
		return nil, fmt.Errorf("error parsing config file %s: %v", configurationPath, err)
	}

	return config, nil
}
