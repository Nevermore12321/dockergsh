package container

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"syscall"

	"github.com/Nevermore12321/dockergsh/utils"
)

// 全局环境变量
var (
	DefaultInfoLocation string = "/var/run/dockergsh/%s/"
	ContainerLogFile    string = "container.log"
)

// container init 进程的信息 结构体
type ContainerInit struct {
	Id       string
	IdBase   string
	ImageUrl string
	RootUrl  string
}


/*
该函数父进程，也就是当前进程执行的内容，
1. ／proc/self/exe 调用，／proc/self／ 指的是当前运行进程自己的环境，exec其实就是自己调用了自己，使用这种方式对创建出来的进程进行初始化
2. 后面的 args 是参数，其中 init 是传递给本进程的第一个参数，在本例中，其实就是会去调用 initCommand去初始化进程的一些环境和资源
3. 下面的 clone 参数就是去 fork 出来一个新进程，并且使用了 namespace 隔离新创建的进程和外部环境。
4. 如果用户指定了 －it 参数，就需要把当前进程的输入输出导入到标准输入输出上

该函数最终返回:
- ContainerInit 容器初始化的结构体
- exec.Cmd 命令结构体
- os.File 一个写管道
*/
func NewParentProcess(tty bool) (*ContainerInit, *exec.Cmd, *os.File) {
	// 初始化管道, 父进程通过管道，将子进程运行的参数传过去
	readPipe, writerPipe, err := utils.NewPipe()
	if err != nil {
		log.Errorf("New pipe err: %v", err)
		return nil, nil, nil
	}

	// 获取当前程序， /proc/self/exec 也就是当前执行的程序
	// 在子进程中执行 /proc/self/exec 也就是子进程执行当前程序
	initCmd, err := os.Readlink("/proc/self/exe")
	if err != nil {
		log.Errorf("get init process error %v", err)
		return nil, nil, nil
	}

	// 通过 os/exec 来 fork 一个子进程并且 执行当前程序，传入 init 参数
	// 也就是在子进程中执行 dockergsh init
	cmd := exec.Command(initCmd, "init")
	//cmd := exec.Command("/proc/self/exe", "init")

	// 在子进程中，添加一个文件描述符. 除了 012， 那么该 readPipe 的文件描述符为 3
	cmd.ExtraFiles = []*os.File{readPipe}
	// 指定 命令的 工作目录
	cmd.Dir = "/root/dockergsh/busybox"
	fmt.Println(cmd, readPipe, writerPipe)

	// 设置 CLONE Flag，（Namespace）
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS | syscall.CLONE_NEWNET | syscall.CLONE_NEWIPC,
	}

	// todo 从镜像构造容器
	id := utils.NewId()
	idBase := utils.Encode([]byte(id))

	// 构造容器的日志
	if tty { // 如果是 -it 选项，那么需要将输入输出都重定向到 标准输入输出
		cmd.Stdout = os.Stdout
		cmd.Stdin = os.Stdin
		cmd.Stderr = os.Stderr
	} else { // 否则，则是 -d 模式，生成容器对应日志目录
		// 创建日志目录，日志目录为  /var/run/dockergsh/contain_id/
		dirURL := fmt.Sprintf(DefaultInfoLocation, id)
		if err := os.MkdirAll(dirURL, 0622); err != nil && os.IsExist(err) {
			log.Errorf("NewParentProcess mkdir %s error %v", dirURL, err)
			return nil, nil, nil
		}

		// 创建日志文件，/var/run/dockergsh/contain_id/container.log
		stdLogFileAbsPath := dirURL + ContainerLogFile
		stdLogFile, err := os.OpenFile(stdLogFileAbsPath, os.O_CREATE|os.O_WRONLY|os.O_SYNC|os.O_APPEND, 0755)
		if err != nil {
			log.Errorf("NewParentProcess create file %s error %v", stdLogFileAbsPath, err)
			return nil, nil, nil
		}

		// 将容器的 输出/错误 重定向到 日志文件
		cmd.Stdout = stdLogFile
		cmd.Stderr = stdLogFile
	}

	// todo 添加 image 信息
	return &ContainerInit{Id: id, IdBase: idBase}, cmd, writerPipe
}

