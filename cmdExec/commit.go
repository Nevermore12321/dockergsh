package cmdExec

import (
	"github.com/Nevermore12321/dockergsh/container"
	log "github.com/sirupsen/logrus"
	"os/exec"
)

func CommitContainer(containerArg, imagesName string) error {
	// 根据用户输入的 容器名称或者容器id，获取容器的 containerInfo
	containerInfo, err := GetContainerInfoByArg(containerArg)
	if err != nil {
		log.Errorf("Get info of %s error %v", containerArg, err)
		return err
	}
	mergeURL := containerInfo.RootUrl + "/" + "merge"
	imageTarURL := container.DefaultFsURL + "images/" + imagesName + ".tar"
	if _, err := exec.Command("tar", "-cvf", imageTarURL, "-C", mergeURL, ".").CombinedOutput(); err != nil {
		log.Errorf("Tar folder %s error %v", mergeURL, err)
		return err
	}
	return nil
}
