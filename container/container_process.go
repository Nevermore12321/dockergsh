package container

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"syscall"
)

/*
创建一个管道
返回：
- 只读管道 - *os.File
- 只写管道 - *os.File
- err - error
 */
func NewPipe() (*os.File, *os.File, error) {
	reader, writer, err := os.Pipe()
	if err != nil {
		return nil, nil, err
	}
	return reader, writer, nil
}
/*
该函数父进程，也就是当前进程执行的内容，
1. ／proc/self/exe 调用，／proc/self／ 指的是当前运行进程自己的环境，exec其实就是自己调用了自己，使用这种方式对创建出来的进程进行初始化
2. 后面的 args 是参数，其中 init 是传递给本进程的第一个参数，在本例中，其实就是会去调用 initCommand去初始化进程的一些环境和资源
3. 下面的 clone 参数就是去 fork 出来一个新进程，并且使用了 namespace 隔离新创建的进程和外部环境。
4. 如果用户指定了 －it 参数，就需要把当前进程的输入输出导入到标准输入输出上
 */
func NewParentProcess() (*exec.Cmd, error) {
	// 初始化管道
	readPipe, writerPipe, err := NewPipe()
	if err != nil {
		log.Errorf("New pipe err: %v", err)
		return nil, nil
	}

	// 获取当前程序， /proc/self/exec 也就是当前执行的程序
	// 在子进程中执行 /proc/self/exec 也就是子进程执行当前程序
	initCmd, err := os.Readlink("/proc/self/exec")
	if err != nil {
		log.Errorf("get init process error %v", err)
		return nil, nil
	}

	// 通过 os/exec 来 fork 一个子进程并且 执行当前程序，传入 init 参数
	cmd := exec.Command(initCmd, "init")
	fmt.Println(cmd, readPipe, writerPipe)

	// 设置 CLONE Flag，（Namespace）
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CloneFlag: syscall.CLONE_NEWUTS,
	}
	return nil,nil
}