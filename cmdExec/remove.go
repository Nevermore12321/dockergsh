package cmdExec

import (
	"fmt"
	"github.com/Nevermore12321/dockergsh/container"
	log "github.com/sirupsen/logrus"
)

func RemoveContainer(containerArg string) error {
	// 获取容器信息
	info, err := GetContainerInfoByArg(containerArg)
	if err != nil {
		log.Errorf("Get container %s info error %v", containerArg, err)
		return err
	}
	// 如果容器正在运行中，不能删除
	if info.Status != container.STOP {
		log.Errorf("Couldn't remove running container")
		return fmt.Errorf("couldn't remove running container")
	}

	// 如果容器已经停止，那么删除容器信息
	deleteContainerInfo(info.Id, info.Name)

	// docker rm 的时候应该已经没有挂载
	mergeURL := info.RootUrl + "/merge"
	container.DeleteWorkSpace(true, "", mergeURL, info.RootUrl)
	return nil
}
