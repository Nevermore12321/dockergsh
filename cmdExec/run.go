package cmdExec

import (
	"encoding/json"
	"fmt"
	"github.com/Nevermore12321/dockergsh/cgroup"
	"github.com/Nevermore12321/dockergsh/cgroup/subsystem"
	"github.com/Nevermore12321/dockergsh/network"
	"github.com/Nevermore12321/dockergsh/utils"
	log "github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/Nevermore12321/dockergsh/container"
)

func Run(tty bool, commandArray []string, resConf *subsystem.ResourceConfig, imageName, containerName, volume string, envSlice []string, networkName string) {
	// containerInit 包含容器初始化时需要记录的一些信息
	// 添加镜像 挂载 等参数
	containerInit, parentCmd, writePipe := container.NewParentProcess(tty, imageName, volume, envSlice)
	if parentCmd == nil { // 如果没有创建出 进程命令
		log.Errorf("New parent process error")
		return
	}
	/*
		这里的 Start 方法是真正开始前面创建好的command的调用:
		1. 首先会 clone 出来一个 namespace 隔离的进程
		2. 然后在子进程中，调用 /proc/self/exe，也就是调用自己，发送 init 参数，调用我们写的 init 方法，去初始化容器的一些资源
		3. 注意，子进程执行 /proc/self/exe ，也就是说要让子进程成为 container 中的 init 程序，需要注意 init 程序不能退出
	*/
	if err := parentCmd.Start(); err != nil {
		log.Errorf("new parent process error: %v", err)
	}

	// record container info
	// 将 Container 详情写入到 文件 config.json 中
	containerName, err := recordContainerInfo(containerInit, parentCmd.Process.Pid, containerName, commandArray, volume)
	if err != nil {
		log.Errorf("Record container info error %v", err)
		return
	}

	fmt.Println(containerInit)

	// 开启cgroup
	// 检查版本
	_, err = exec.Command("grep", "cgroup2", "/proc/filesystems").CombinedOutput()
	if err != nil { // cgroup v1
		// use dockergsh as cgroup name
		cgroupManager := cgroup.NewCgroupManager(containerInit.IdBase)
		//  如果以 -it 启动容器，那么退出时，直接删除 cgroup
		if tty {
			defer cgroupManager.DestroyV1()
		}
		// 设置资源限制
		if err := cgroupManager.SetV1(resConf); err != nil {
			log.Errorf("set cgroup resource failed: %v", err)
		}
		// 将容器进程 pid 加入到 cgroup 中
		if err = cgroupManager.ApplyV1(parentCmd.Process.Pid); err != nil {
			log.Errorf("add process to cgroup failed: %v", err)
		}
	} else { // cgroup v2
		// use dockergsh as cgroup name
		cgroupManager := cgroup.NewCgroupManager(containerInit.IdBase)
		//  如果以 -it 启动容器，那么退出时，直接删除 cgroup
		if tty {
			defer cgroupManager.DestroyV2()
		}
		// 设置资源限制
		if err := cgroupManager.SetV2(resConf); err != nil {
			log.Errorf("set cgroup resource failed: %v", err)
		}
		// 将容器进程 pid 加入到 cgroup 中
		if err = cgroupManager.ApplyV2(parentCmd.Process.Pid); err != nil {
			log.Errorf("add process to cgroup failed: %v", err)
		}
	}

	// 配置容器网络
	if networkName != "" {
		err := network.Init()
		if err != nil {
			log.Errorf("network init failed: %v", err)
		}
		// todo 端口映射
		containerInfo := container.ContainerInfo{
			Id:   containerInit.Id,
			Pid:  strconv.Itoa(parentCmd.Process.Pid),
			Name: containerName,
			//PortMapping: portmapping,
		}

		if err = network.ConnectNetwork(networkName, &containerInfo); err != nil {
			log.Errorf("Error Connect Network %v", err)
			return
		}
	}

	// 父进程向容器中发送 所有的命令选项
	sendInitCommand(commandArray, writePipe)

	// 如果是 -it 伪终端模式，那么需要监听，如果退出，需要释放容器资源
	if tty {
		// parent.Wait() 主要是用于父进程等待子进程结束
		if err := parentCmd.Wait(); err != nil {
			log.Errorf("Wait for child err: %v", err)
		}

		// 容器删除后，删除容器的记录信息
		deleteContainerInfo(containerInit.Id, containerName)

		// todo 停止容器时，删除挂载路径
		container.DeleteWorkSpace(true, volume, containerInit.MergeUrl, containerInit.RootUrl)
	}
}

// 向管道中发送消息
// 也就是父进程通过管道向子进程（容器）中发送命令行选项
func sendInitCommand(comArray []string, writePipe *os.File) {
	// 把所有选项通过空格分割
	commandOpt := strings.Join(comArray, " ")
	log.Infof("Send command to init container: %s", commandOpt)
	_, err := writePipe.WriteString(commandOpt)
	if err != nil {
		log.Warnf("Send command Opt to container init failed: %s", err)
	}
	err = writePipe.Close()
	if err != nil {
		log.Warnf("Pipe close failed: %s", err)
	}
}

/*
记录容器的信息
将 container 的详细信息写入到 /var/lib/dockergsh/[containerID]/container/config.json
*/
func recordContainerInfo(containerInit *container.ContainerInit, containerPid int, containerName string, commandArray []string, volume string) (string, error) {
	//  创建时间
	createTime := time.Now().Format("2006-01-02 15:04:05")
	//  容器的启动命令
	command := strings.Join(commandArray, " ")
	log.Infof("Container command is %s:", command)
	var idFlag bool
	// 如果 docker 启动的时候没有指定名称，那么就是用 id
	if containerName == "" {
		containerName = containerInit.Id
		idFlag = true
	}

	// 初始化 ConntainerInfo 实例
	containerInfo := &container.ContainerInfo{
		Name:       containerName,
		Pid:        strconv.Itoa(containerPid),
		Id:         containerInit.Id,
		Command:    command,
		CreateTime: createTime,
		Status:     container.RUNNING,
		RootUrl:    containerInit.RootUrl,
		Volume:     volume,
	}

	// 将 ContainerInfo 结构体实例 转成 json 字符串
	jsonBytes, err := json.Marshal(containerInfo)
	if err != nil {
		log.Errorf("Record container info error %v", err)
		return "", err
	}
	jsonStr := string(jsonBytes)

	//  如果 目录没有创建，则创建
	configFileURL := containerInfo.RootUrl + "/" + container.ContainerConfigPath + "/"
	if err := os.MkdirAll(configFileURL, 0622); err != nil {
		log.Errorf("Mkdir error %s error %v", configFileURL, err)
		return "", err
	}

	// 如果 idFlag 为 false，也就是 docker 启动时设置了容器的名称，那么就添加一个软链接 /var/lib/dockergsh/named_containers/[containerName] 到 /var/lib/dockergsh/[containerID]
	// 便于观察
	if !idFlag {
		containersUrl := fmt.Sprintf(container.DefaultInfoLocation, container.NamedContainersDir)
		if exist, err := utils.PathExists(containersUrl); err != nil {
			log.Errorf("Soft link floder %s create err: %v", containersUrl, err)
			return "", err
		} else if !exist {
			if err := os.MkdirAll(containersUrl, 0777); err != nil {
				log.Errorf("Create Soft Link Foldeer Failed:  %s . %v", containersUrl, err)
				return "", err
			}
		}
		linkURL := containersUrl + containerName
		if err := os.Symlink(configFileURL, linkURL); err != nil {
			log.Errorf("Soft link error %s error %v", configFileURL, err)
			return "", err
		}
	}

	// 创建 config.json 文件
	configFilePath := configFileURL + container.ConfigName
	file, err := os.Create(configFilePath)
	defer file.Close()
	if err != nil {
		log.Errorf("Create file %s error %v", configFilePath, err)
		return "", err
	}

	// 将 container 详情写入 config.json 文件
	if _, err := file.WriteString(jsonStr); err != nil {
		log.Errorf("File write string error %v", err)
		return "", err
	}

	return containerName, nil
}

/*
删除容器的信息
*/
func deleteContainerInfo(containerId, containerName string) {
	// 删除 /var/lib/dockergsh/[containerId]/container 目录
	rootDir := fmt.Sprintf(container.DefaultInfoLocation, containerId)
	configFileURL := rootDir + "container"
	if err := os.RemoveAll(configFileURL); err != nil {
		log.Errorf("Remove dir %s error %v", configFileURL, err)
	}

	// 删除 软链接
	linkUrl := fmt.Sprintf(container.DefaultInfoLocation[:len(container.DefaultInfoLocation)], "named_containers/"+containerName)
	if err := os.RemoveAll(linkUrl); err != nil {
		log.Errorf("Remove dir %s error %v", linkUrl, err)
	}
}
