package container

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

/*
这里的 init 函数是在容器内部执行的，也就是说，代码执行到这里后，容器所在的进程其实就已经创建出来了，这是本容器执行的第一个进程。
使用 mount 先去挂载 proc 文件系统，以便后面通过 ps 等系统命令去查看当前进程资源的情况。
*/
func RunContainerInitProcess() error {
	// 容器的初始化 init 进程

	//  读取传入的命令
	cmdArray := readUserCommand()
	if cmdArray == nil || len(cmdArray) == 0 {
		return fmt.Errorf("Run container get user command")
	}

	// 获取子进程执行的 dockergsh 程序的绝对路径
	// 这个函数帮我们在当前系统的PATH里面去寻找命令的绝对路径，然后运行起来。
	cmdPath, err := exec.LookPath(cmdArray[0])
	if err != nil {
		log.Errorf("Exec loop path error %v", err)
		return err
	}
	log.Infof("Find path %s", cmdPath)

	// todo 设置挂载点, mount proc 文件系统
	//setUpMount()

	// 这里的 MountFlag 的意思如下:
	// 1. MS_NOEXEC - 在本文件系统中不允许运行其他程序。
	// 2. MS_NOSUID - 在本系统中运行程序的时候，不允许 set-user-ID 或 set-group-ID
	// 3. MS_NODEV - 这个参数是自从 Linux2.4 以来，所有 mount 的系统都会默认设定的参数。本函数最后的s y s c a l l . E x e c，是最为重要的一句黑魔法，正是这个系统调用实现了完成
	defaultMountFlags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV
	syscall.Mount("proc", "/proc", "proc", uintptr(defaultMountFlags), "")

	// 使用 syscall.Exec 执行命令
	if err = syscall.Exec(cmdPath, cmdArray[0:], os.Environ()); err != nil {
		log.Errorf(err.Error())
	}

	return nil
}

/*
子进程，也就是 container init 进程，通过 pipe 管道读取命令选项
在通过 namespace 隔离后，文件描述符也被隔离，因此 在 container 子进程中，
1 是标准输出（stdout）
2 是标准错误输出（stderr）
0 是标准输入（stdin）
那么 3 就是在传入子进程的 文件描述符
*/
func readUserCommand() []string {
	log.Infof("Read parent pipe cmd.")
	// 打开 管道
	pipe := os.NewFile(uintptr(3), "pipe")
	// 从管道中读取 命令选项
	msg, err := ioutil.ReadAll(pipe)
	if err != nil {
		log.Errorf("Init read pipe error %v", err)
		return nil
	}
	msgStr := string(msg)
	log.Infof("receive %s", msgStr)
	return strings.Split(msgStr, " ")
}

/**
Init 挂载点
*/
func setUpMount() {
	pwd, err := os.Getwd()
	if err != nil {
		log.Errorf("Get current working directory error. %s", err)
	}
	log.Infof("Current location is [%s]", pwd)

	//  mount -t proc proc /proc
	// syscall.Mount(source string, target string, fstype string, flags uintptr, data string)
	syscall.Mount("", "/", "", syscall.MS_PRIVATE|syscall.MS_REC, "")
	// todo
}

/**
为了使当前root的老 root 和新 root 不在同一个文件系统下，我们把root重新mount了一次
bind mount是把相同的内容换了一个挂载点的挂载方法
如果不做这一步，就会让其他root没有了 proc 文件系统
*/
func pivotRoot(root string) error {
	// mount root system 的文件系统
	if err := syscall.Mount(root, root, "bind", syscall.MS_BIND|syscall.MS_REC, ""); err != nil {
		return fmt.Errorf("Mount rootfs to itself error: %v", err)
	}

	// 创建 rootfs/.pivot_root 存储 old_root, 也就是 ~/.pivot_root
	pivotDir := filepath.Join(root, ".pivot_root")
	if err := os.Mkdir(pivotDir, 0777); err != nil {
		return err
	}
	// todo
}
