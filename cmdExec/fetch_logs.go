package cmdExec

import (
	"fmt"
	"github.com/Nevermore12321/dockergsh/container"
	"github.com/Nevermore12321/dockergsh/utils"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"path/filepath"
)

func LogContainer(containerArg string) error {
	// 通过用户传入的 容器名称或者 容器id 获取 容器id
	containerInfo, err := GetContainerInfoByArg(containerArg)
	if err != nil {
		logrus.Errorf("Get Container %s Info err error %v", containerArg, err)
		return err
	}

	logFileLocation := containerInfo.RootUrl + "/" + container.ContainerLogFile
	file, err := os.Open(logFileLocation)
	if err != nil {
		logrus.Errorf("Log container open file %s error %v", logFileLocation, err)
		return err
	}

	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		if err != nil {
			logrus.Errorf("Log container read file %s error %v", logFileLocation, err)
			return err
		}
	}
	_, err = fmt.Fprint(os.Stdout, string(content))
	if err != nil {
		logrus.Errorf("Log container Print file error %v", err)
		return err
	}
	return nil
}

// GetContainerInfoByArg 用户在使用时，可以传入容器名称，也可以传入容器的id
// GetContainerInfoByArg 通过判断是容器名还是 容器id，返回查到的 容器信息  ContainerInfo
func GetContainerInfoByArg(containerArg string) (*container.ContainerInfo, error) {
	var containerInfo *container.ContainerInfo

	// 先在 named_containers 目录下找有没有容器，判断该容器有没有命名
	// 表示有名字的容器
	namedContainersUrl := fmt.Sprintf(container.DefaultInfoLocation, container.NamedContainersDir)
	files, err := os.ReadDir(namedContainersUrl)
	if err != nil {
		logrus.Errorf("Read container base folder %s error %v", namedContainersUrl, err)
		return nil, err
	}
	// idFlag 表示 containerArg 表示的是否是 id
	for _, file := range files {
		// 如果是 containerName
		if file.Name() == containerArg {
			configURL := filepath.Join(namedContainersUrl, file.Name(), container.ConfigName)
			// 读取配置文件 config.json
			containerInfo, err = GetContainerInfo(configURL)
			if err != nil {
				if err != os.ErrInvalid {
					logrus.Errorf("Get container info by name error %v", err)
					return nil, err
				}
			}
			break
		}
	}
	if containerInfo == nil {
		hashId := utils.EncodeSha256([]byte(containerArg))
		containerURL := fmt.Sprintf(container.DefaultInfoLocation, hashId)
		configURL := filepath.Join(containerURL, container.ContainerConfigPath, container.ConfigName)
		containerInfo, err = GetContainerInfo(configURL)
		if err != nil {
			if err != os.ErrInvalid {
				logrus.Errorf("Get container info by id error %v", err)
				return nil, err
			}
		}
	}

	return containerInfo, nil

}
