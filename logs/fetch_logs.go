package logs

import (
	"fmt"
	"github.com/Nevermore12321/dockergsh/cmdExec"
	"github.com/Nevermore12321/dockergsh/container"
	"github.com/Nevermore12321/dockergsh/utils"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"path/filepath"
)

func LogContainer(containerArg string) error {
	// 先在 named_containers 目录下找有没有容器，判断该容器有没有命名
	// 表示有名字的容器
	containersUrl := fmt.Sprintf(container.DefaultInfoLocation, container.NamedContainersDir)
	files, err := os.ReadDir(containersUrl)
	// idFlag 表示 containerArg 表示的是否是 id
	for _, file := range files {
		// 如果是 containerName
		if file.Name() == containerArg {
			configURL := filepath.Join(containersUrl, file.Name(), container.ConfigName)
			// 读取配置文件 config.json
			info, err := cmdExec.GetContainerInfo(configURL)
			if err != nil {
				if err != os.ErrInvalid {
					logrus.Errorf("Get container info error %v", err)
				}
			}
			containerArg = info.Id
			break
		}
	}

	hashId := utils.EncodeSha256([]byte(containerArg))

	dirURL := fmt.Sprintf(container.DefaultInfoLocation, hashId)
	logFileLocation := dirURL + container.ContainerLogFile
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
