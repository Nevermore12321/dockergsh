package iptables

import (
	"errors"
	"fmt"
	"github.com/Nevermore12321/dockergsh/internal/client"
	"os"
	"os/exec"
	"strings"
)

var (
	ErrIptablesNotFound = errors.New("iptables not found")
	supportsXlock       = false
)

// Exists 判断当前 itables 规则是否已经存在，args 表示 iptables 后面跟的参数
func Exists(args ...string) bool {
	if _, err := Raw(append([]string{"-C"}, args...)...); err != nil {
		return false
	}
	return true
}

// Raw 执行 iptables 命令，args 表示 iptables 后面跟的参数
func Raw(args ...string) ([]byte, error) {
	// 判断是否存在 iptables 命令
	path, err := exec.LookPath("iptables")
	if err != nil {
		return nil, ErrIptablesNotFound
	}

	if supportsXlock {
		args = append([]string{"--wait"}, args...)
	}

	// debug 信息
	if os.Getenv(client.DOCKERGSH_DEBUG) != "" {
		fmt.Fprintf(os.Stderr, fmt.Sprintf("[debug] %s, %v \n", path, args))
	}

	// 执行 iptables 命令
	output, err := exec.Command(path, args...).CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("iptables failed: iptables %v: %s (%s)", strings.Join(args, " "), output, err)
	}

	// 忽略 iptables 关于 xtables 锁的消息
	if strings.Contains(string(output), "waiting for it to exit") {
		output = []byte("")
	}
	return output, nil
}
