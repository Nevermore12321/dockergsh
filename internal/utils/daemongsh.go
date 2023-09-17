package utils

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
	"strconv"
)

// CreatePidFile 创建一个 pid 文件
func CreatePidFile(pidFile string) error {
	// 读取pid文件
	pidContent, err := os.ReadFile(pidFile)
	if err == nil { // 如果已经存在了 pid 文件，在判断是否运行了 pid 号进程
		// 文件内容就是 pid 进程号
		pid, err := strconv.Atoi(string(pidContent))
		if err == nil {
			// 检查是否有 pid 号进程运行文件
			if _, err := os.Stat(fmt.Sprintf("/proc/%d/", pid)); err == nil {
				// 如果有 pid 运行文件，报错返回
				return fmt.Errorf("pid file found, ensure docker is not running or delete %s", pidfile)
			}
		}
	}

	// 如果没有存在 pid 文件，那么就需要创建
	file, err := os.Create(pidFile)
	if err != nil {
		return err
	}

	defer file.Close()

	// 向 pid 文件中写入当前进程 pid 进程号
	_, err = fmt.Fprintf(file, "%d", os.Getpid())
	return err
}

// RemovePidFile 删除一个 pid 文件
func RemovePidFile(pidFile string) {
	if err := os.Remove(pidFile); err != nil {
		log.Printf("Error removing %s: %s", pidFile, err)
	}
}
