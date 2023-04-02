package cmdExec

import (
	"encoding/json"
	"github.com/Nevermore12321/dockergsh/container"
	"github.com/sirupsen/logrus"
	"os"
	"path/filepath"
	"strconv"
	"syscall"
)

func StopContainer(containerArg string) error {
	// 根据用户输入的 containerId 或者 containerName 获取 contianer Info
	info, err := GetContainerInfoByArg(containerArg)
	if err != nil {
		logrus.Errorf("Get Container %s Info err error %v", containerArg, err)
		return err
	}
	containerPid, err := strconv.Atoi(info.Pid)
	if err != nil {
		logrus.Errorf("Conver pid from string to int error %v", err)
		return err
	}

	// 系统调用 kill，发送 SigTerm 信号给进程，杀掉容器的主进程
	if err = syscall.Kill(containerPid, syscall.SIGTERM); err != nil {
		logrus.Errorf("Stop container %s error %v", info.Id, err)
		return err
	}

	// 修改容器状态，将 Running 改为 Stopped，pid 可以设置为空
	info.Pid = ""
	info.Status = container.STOP
	if err = UpdateContainerInfo(info); err != nil {
		logrus.Errorf("Update container info  %s error, %v", info.Id, err)
		return err
	}
	return nil

}

// UpdateContainerInfo 根据 info 中的 容器 id 找到对应的 container 信息，并且修改
func UpdateContainerInfo(info *container.ContainerInfo) error {
	// 将 containerInfo 序列化成 json 字符串
	infoBytes, err := json.Marshal(info)
	if err != nil {

		logrus.Errorf("Json marshal %s error %v", info.Id, err)
		return err
	}
	configFilePath := filepath.Join(info.RootUrl, container.ContainerConfigPath, container.ConfigName)
	if err = os.WriteFile(configFilePath, infoBytes, 0622); err != nil {
		logrus.Errorf("Write file %s error, %v", configFilePath, err)
		return err
	}
	return nil
}
