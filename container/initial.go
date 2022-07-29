package container

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	log "github.com/sirupsen/logrus"
)

var old_root = ".pivot_root"

/*
这里的 init 函数是在容器内部执行的，也就是说，代码执行到这里后，容器所在的进程其实就已经创建出来了，这是本容器执行的第一个进程。
使用 mount 先去挂载 proc 文件系统，以便后面通过 ps 等系统命令去查看当前进程资源的情况。
*/
func RunContainerInitProcess() error {
	// 容器的初始化 init 进程

	// 设置挂载点, mount proc 文件系统
	setUpMount()

	//  读取传入的命令
	cmdArray := readUserCommand()
	if cmdArray == nil || len(cmdArray) == 0 {
		return fmt.Errorf("Run container get user command")
	}

	// 获取子进程执行的 dockergsh 程序的绝对路径
	// 这个函数帮我们在当前系统的PATH里面去寻找命令的绝对路径，然后运行起来。
	// LookPath("pwd") 也就是判断 pwd 命令的绝对路径存不存在
	cmdPath, err := exec.LookPath(cmdArray[0])
	if err != nil {
		log.Errorf("Exec loop path error %v", err)
		return err
	}
	log.Infof("Find path %s", cmdPath)

	// 使用 syscall.Exec 执行命令, 执行 docker run 最后跟的命令
	if err = syscall.Exec(cmdPath, cmdArray[0:], os.Environ()); err != nil {
		log.Infof("cmdPath: %v", cmdPath)
		log.Errorf(err.Error())
		return err
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
	// 获取当前路径
	pwd, err := os.Getwd()
	if err != nil {
		log.Errorf("Get current working directory error. %s", err)
	}
	log.Infof("Current location is [%s]", pwd)

	// 解决  pivot_root 调用 Invalid arguments 错误
	// 原因 pivot root 不允许 parent mount point 和 new mount point 是 shared。
	syscall.Mount("", "/", "", syscall.MS_PRIVATE|syscall.MS_REC, "")

	if err := pivotRoot(pwd); err != nil {
		log.Errorf("Error when call pivotRoot %v", err)
	}

	//  mount -t proc proc /proc
	//  mount -t tmpfs tmpfs /dev ： tmpfs是一种基于内存的文件系统，可以使用RAM或swap分区来存储。
	// syscall.Mount(source string, target string, fstype string, flags uintptr, data string)
	// 这里的 MountFlag 的意思如下:
	// 1. MS_NOEXEC - 在本文件系统中不允许运行其他程序。
	// 2. MS_NOSUID - 在本系统中运行程序的时候，不允许 set-user-ID 或 set-group-ID
	// 3. MS_NODEV - 这个参数是自从 Linux2.4 以来，所有 mount 的系统都会默认设定的参数。本函数最后的s y s c a l l . E x e c，是最为重要的一句黑魔法，正是这个系统调用实现了完成
	defaultMountFlags := syscall.MS_NOEXEC | syscall.MS_NODEV | syscall.MS_NOSUID
	syscall.Mount("proc", "/proc", "proc", uintptr(defaultMountFlags), "")
	syscall.Mount("tmpfs", "/dev", "tmpfs", syscall.MS_NOSUID|syscall.MS_STRICTATIME, "mode=755")

}

/**
为了使当前root的老 root 和新 root 不在同一个文件系统下，我们把root重新mount了一次
bind mount是把相同的内容换了一个挂载点的挂载方法
如果不做这一步，就会让其他root没有了 proc 文件系统
- 通过 pivot_root 将 root 的文件系统一道 put_old
- umount root
- mount new_root
- mount old root
*/
func pivotRoot(root string) error {
	// mount root system 的文件系统
	// new_root 和 put_old 必须不能同时存在当前root 的同一个文件系统中,需要通过--bind重新挂载一下
	// 为了使当前 root 的老 root 和新 root 不在同一个文件系统下，把 root 重新 mount 了一次
	// bind mount 是把相同的内容换了一个挂载点的挂载方法
	if err := syscall.Mount(root, root, "bind", syscall.MS_BIND|syscall.MS_REC, ""); err != nil {
		return fmt.Errorf("mount --bind PWD PWD error : %v", err)
	}

	// 创建 rootfs/.pivot_root 存储 old_root，类似 ~/.pivot_root 文件夹
	pivotDir := filepath.Join(root, old_root)
	if err := os.Mkdir(pivotDir, 0777); err != nil && !os.IsExist(err) {
		log.Errorf("Failed to create putOld folder %s, error: %v", pivotDir, err)
		return err
	}

	// 通过 pivot_root 系统调用，把root的老 old_root 挂载到rootfs/.pivot_root
	// 挂载点现在依然可以在 mount 命令中查看
	if err := syscall.PivotRoot(root, pivotDir); err != nil {
		return fmt.Errorf("pivot_root %v", err)
	}

	// 修改容器的工作目录
	if err := syscall.Chdir("/"); err != nil {
		return fmt.Errorf("chdir / %v", err)
	}

	// 因为通过 pivot_root 系统调用将原本的 root 挂载到了rootfs/.pivot_root , 也就是 old_root
	// 现在需要将 old_root 再次解除挂载，因为之前有 mount root --bind
	// 1. 将当前目录 mount --bind pwd pwd
	// 2. pivot_root 将容器的 root 挂载，然后将老的 root 放到 rootfs/.pivot_root （容器进程，MOUNT Namespace 隔离）
	// 3. 解除 rootfs/.pivot_root 挂载，因为有 mount --bind 第一步，因此，解除 rootfs/.pivot_root 原本的 root 文件系统挂载还在
	pivotDir = filepath.Join("/", old_root)
	if err := syscall.Unmount(pivotDir, syscall.MNT_DETACH); err != nil {
		return fmt.Errorf("unmount pivot_root dir %v", err)
	}

	// 删除临时文件夹
	return os.Remove(pivotDir)
}
