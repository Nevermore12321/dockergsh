package container

import (
	log "github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"strings"
)

func CreateVolume(volume, mergeURL string) error {
	if volume != "" {
		// 提取 volume 挂载信息 xxx:xxx -> [xxx,xxx]
		volumeURLs := volumeUrlExtract(volume)
		length := len(volumeURLs)
		// 必须满足 xxx:xxx 格式
		if length == 2 && volumeURLs[0] != "" && volumeURLs[1] != "" {
			//  挂载 volume
			if err := MountVolume(volumeURLs, mergeURL); err != nil {
				return err
			}
			log.Infof("%q", volumeURLs)
		} else {
			log.Warnf("Volume parameter input is not correct")
		}
	}
	return nil
}

func DeleteVolume(volume, mergeURL string) error {
	if volume != "" {
		volumeURLs := volumeUrlExtract(volume)
		length := len(volumeURLs)
		if length == 2 && volumeURLs[0] != "" && volumeURLs[1] != "" {
			_ = DeleteVolumeMountPoint(volumeURLs, mergeURL)
		}
	}
	return nil
}

/*
docker 挂载 volume 的格式为  docker run -v hostPath:containerPath xxx
volumeUrlExtract 函数是将 参数 volume xxx:xxx 分成 [xxx, xxx]
 */
func volumeUrlExtract(volume string) []string {
	return strings.Split(volume, ":")
}

/*
挂载 volume 的实际操作
 */
func MountVolume(volumeURLs []string, mergeURL string) error  {
	// 判断 宿主机需要挂载目录是否存在，不存在直接创建
	hostURL := volumeURLs[0]
	if err := os.MkdirAll(hostURL, 0777); err != nil {
		log.Infof("Mkdir host volume dir %s error. %v", hostURL, err)
		return err
	}
	// 判断 容器中挂载目录是否存在，不存在直接创建
	// 这里需要注意，容器挂载目录，是需要在 merge 层之上创建的
	// 例如 docker run -v /home/gsh:/home/container
	// 其实是把宿主机的 /home/gsh 挂载到 /var/lib/dockergsh/[containerID]/merge/home/container
	containerURL := volumeURLs[1]
	containerVolumeURL := mergeURL + "/" + containerURL
	if err := os.MkdirAll(containerVolumeURL, 0777); err != nil {
		log.Infof("Mkdir container volume dir %s error. %v", containerVolumeURL, err)
		return err
	}

	//  mount 挂载
	mountCmd := exec.Command("mount", "--bind", hostURL, containerVolumeURL)
	mountCmd.Stdout = os.Stdout
	mountCmd.Stderr = os.Stderr
	if err := mountCmd.Run(); err != nil {
		log.Errorf("Mount volume failed. %v", err)
		return err
	}
	return nil
}


/*
解除 volume 的挂载
 */
func DeleteVolumeMountPoint(volumeURLs []string, mergeURL string) error {
	// 容器中的 volume 实际目录在 merge layer 下
	containerURL := mergeURL + "/" + volumeURLs[1]

	// umount volume
	umountCmd := exec.Command("umount", containerURL)
	umountCmd.Stderr = os.Stderr
	umountCmd.Stdout = os.Stdout
	if err := umountCmd.Run(); err != nil {
		log.Errorf("Umount volume %s failed. %v", containerURL, err)
		return err
	}
	return nil
}