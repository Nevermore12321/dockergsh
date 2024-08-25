//go:build linux

package v1

import (
	"bufio"
	"os"
	"strings"
)

// FindCgroupMountPoint 通过 /proc/self/mountinfo 找到挂载了某个 subsystem 的 hierarchy cgroup 的根节点所在目录
//
//	使用：FindCgroupMountPoint("memory"),这里返回具体某个 cgroup 挂载的根路径
func FindCgroupMountPoint(subsystem string) (string, error) {
	// 打开 /proc/self/mountinfo
	file, err := os.Open("/proc/self/mountinfo")
	if err != nil {
		return "", err
	}
	defer file.Close()
	// 读取 /proc/self/mountinfo 文件内容
	// mountinfo 文件都是以空格分开的，因此以空格分割后，找到第五个字段就是挂载的路径
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		text := scanner.Text()
		splitFields := strings.Split(text, " ")
		// 首先判断是什么 subsystem 类型，也就是最后一个字段 rw,memory
		for _, value := range strings.Split(splitFields[len(splitFields)-1], ",") {
			if value == subsystem {
				return splitFields[4], nil
			}
		}
	}
	return "", nil
}
