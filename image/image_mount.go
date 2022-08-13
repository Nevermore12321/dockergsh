package image

import (
	"fmt"
	"github.com/Nevermore12321/dockergsh/utils"
	log "github.com/sirupsen/logrus"
	"os"
	"os/exec"
)

var (
	DefaultImageDir = "/var/lib/dockergsh/images/"
	DefaultImageLowerLayer = "/var/lib/dockergsh/images/busybox/"
)

// 创建容器的 根目录
func CreateRootDir(rootURL string) error {
	exists, err := utils.PathExists(rootURL)
	if err != nil {
		log.Infof("Fail to judge whether dir %s exists. %v", rootURL, err)
		return err
	}
	if !exists {
		if err := os.Mkdir(rootURL, 0777); err != nil {
			log.Errorf("Create RootDir Failed:  %s . %v", rootURL, err)
			return err
		}
	}
	return nil
}

/*
创建 image 层 ，也就是不可修改层
也就是 pull 下来的镜像，保存的位置，就是 imageURL
将 tar 格式的镜像解压 ，其实就是镜像层的，不可修改层 layer
*/
func CreateLowerLayer(imageURL, rootURL string) error {
	if imageURL == "" {
		return fmt.Errorf("image URl is nil")
	}

	// imageTarURL 镜像 tar 格式的位置
	// imageMountURL docker 把 lower layer 保存的具体位置，也就是 rootURL/lower
	var imageTarURL, imageLowerLayerURL string
	imageTarURL = DefaultImageDir + imageURL
	imageLowerLayerURL = rootURL + "/lower"

	exists, err := utils.PathExists(imageLowerLayerURL)
	if err != nil {
		log.Infof("Fail to judge whether dir %s exists. %v", imageLowerLayerURL, err)
		return err
	}
	if !exists {
		// 创建 lower layer 文件夹
		if err := os.MkdirAll(imageLowerLayerURL, 0777); err != nil {
			log.Errorf("Mkdir Image Lower Layer %s error. %v", imageLowerLayerURL, err)
			return err
		}

		// 解压 tar 格式的镜像, 解压到 lower layer 的目录中
		if _, err := exec.Command("tar", "-xvf", imageTarURL, "-C", imageLowerLayerURL).CombinedOutput(); err != nil {
			log.Errorf("unTar tar format image form dir %s error. %v", imageTarURL, err)
			return err
		}
	}
	return nil
}

/*
创建 镜像层 lower 之上的 upper 层，作为读写层
其实就是 Container 层，在启动一个容器的时候会在最后的image层的上一层自动创建，所有对容器数据的更改都会发生在这一层
 */
func CreateUpperLayer(rootURL string) error {
	upperLayerURL := rootURL + "/upper"
	exists, err := utils.PathExists(upperLayerURL)
	if err != nil {
		log.Infof("Fail to judge whether dir %s exists. %v", upperLayerURL, err)
		return err
	}
	// 如果不存在 upper layer 目录，就创建
	if !exists {
		if err := os.MkdirAll(upperLayerURL, 0777); err != nil {
			log.Errorf("Mkdir Upper Layer dir %s error. %v", upperLayerURL, err)
			return err
		}
	}
	return nil
}

/*
创建 work 文件夹作为最终的挂载点
 */
func CreateWorkDir(rootURL string) error {
	workLayerURL := rootURL + "/work"
	exists, err := utils.PathExists(workLayerURL)
	if err != nil {
		log.Infof("Fail to judge whether dir %s exists. %v", workLayerURL, err)
		return err
	}
	// 如果不存在 worker layer 目录，就创建
	if !exists {
		if err := os.MkdirAll(workLayerURL, 0777); err != nil {
			log.Errorf("Mkdir Upper Layer dir %s error. %v", workLayerURL, err)
			return err
		}
	}
	return nil
}

/*
创建 merge layer
lower、upper、worker 三种目录合并出来的目录，merged 目录里面本身并没有任何实体文件
这里的目录，其实就是最终容器的工作目录
 */
func CreateMountPoint(imageURL, mergeURL, rootURL string) error {
	mergeLayerURL := mergeURL
	exists, err := utils.PathExists(mergeLayerURL)
	if err != nil {
		log.Infof("Fail to judge whether dir %s exists. %v", mergeLayerURL, err)
		return err
	}
	// 如果不存在 merge layer 目录，就创建
	if !exists {
		if err := os.MkdirAll(mergeLayerURL, 0777); err != nil {
			log.Errorf("Mkdir Upper Layer dir %s error. %v", mergeLayerURL, err)
			return err
		}
	}

	var mountDirs string
	// 这里是使用 overlay2 将 lower、upper、worker 三个目录，挂载至 rootURL/merge 目录
	// 如果 imageURL 不存在，使用一个默认的文件系统 busybox 镜像
	if imageURL != "" {
		mountDirs = "lowerdir=" + rootURL + "/lower" +
			",upperdir=" + rootURL + "/upper" +
			",workdir=" + rootURL + "/work"
	} else {
		mountDirs = "lowerdir=" + DefaultImageLowerLayer +
			",upperdir=" + rootURL + "/upper" +
			",workdir=" + rootURL + "/work"
	}

	mountCmd := exec.Command("mount", "-t", "overlay", "overlay", "-o", mountDirs, mergeLayerURL)
	mountCmd.Stdout = os.Stdout
	mountCmd.Stderr = os.Stderr

	if err := mountCmd.Run(); err != nil {
		log.Errorf("Mount overlay to merge failed: %v", err)
		return err
	}
	return nil
}

/*
删除读写层，也就是 container 的根目录
 */
func DeleteWriteLayer(rootURL string) error {
	if err := os.RemoveAll(rootURL); err != nil {
		log.Errorf("Remove WriteLayer dir %s error %v", rootURL, err)
		return err
	}
	return nil
}

/*
解除挂载 merge layer
 */
func DeleteMountPoint(mergeURL string) error {
	umountCmd := exec.Command("umount", mergeURL)
	umountCmd.Stderr = os.Stderr
	umountCmd.Stdout = os.Stdout
	if err := umountCmd.Run(); err != nil {
		log.Errorf("Unmount merge layer %s error %v", mergeURL, err)
		return err
	}

	return nil
}