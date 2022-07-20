package cmdExec

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
	"strings"

	"github.com/Nevermore12321/dockergsh/container"
)

func Run(tty bool, commandArray []string) {
	// containerInit 包含容器初始化时需要记录的一些信息
	// todo 添加镜像 挂载 等参数
	containerInit, parentCmd, writePipe := container.NewParentProcess(tty)
	if parentCmd == nil {			// 如果没有创建出 进程命令
		log.Errorf("New parent process error")
		return
	}
	/*
	这里的 Start 方法是真正开始前面创建好的command的调用:
	1. 首先会 clone 出来一个 namespace 隔离的进程
	2. 然后在子进程中，调用／proc/self/exe，也就是调用自己，发送 init 参数，调用我们写的 init 方法，去初始化容器的一些资源
	3. 注意，子进程执行 ／proc/self/exe ，也就是说要让子进程成为 container 中的 init 程序，需要注意 init 程序不能退出
	 */
	if err:= parentCmd.Start(); err != nil {
		log.Errorf("New parent process error: %v", err)
	}

	// todo
	// record container info
	// 开启cgroup
	// 检查版本
	fmt.Println(containerInit)

	// 父进程向容器中发送 所有的命令选项
	sendInitCommand(commandArray, writePipe)


	// 如果是 -it 伪终端模式，那么需要监听，如果退出，需要释放容器资源
	if tty {
		if err := parentCmd.Wait(); err != nil {
			log.Errorf("Wait for child err: %v", err)
		}

		// todo 停止容器
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