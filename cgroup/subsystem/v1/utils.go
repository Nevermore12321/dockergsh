package v1

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"strings"
)

// 通过 /proc/self/mountinfo 找到挂载了某个 subsystem 的 hierarchy cgroup 的根节点所在目录
//  使用：FindCgroupMountPoint("memory"),这里返回具体某个 cgroup 挂载的根路径
func FindCgroupMountPoint(subsystem string) string {
	// 打开 /proc/self/mountinfo
	file, err := os.Open("/proc/self/mountinfo")
	if err != nil {
		return ""
	}
	defer file.Close()
	// 读取 /proc/self/mountinfo 文件内容
	// mountinfo 文件都是以空格分开的，因此以空格分割后，找到第五个字段就是挂载的路径
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		text := scanner.Text()
		split_fields := strings.Split(text, " ")
		// 首先判断是什么 subsystem 类型，也就是最后一个字段 rw,memory
		for _, value := range strings.Split(split_fields[len(split_fields) - 1], ",") {
			if value == subsystem {
				return split_fields[4]
			}
		}
	}
	return ""
}


// GetCgroupPath 函数是找到对应 subsystem 挂载的相对路径，然后通过对这个目录的读写取操作 cgroup
// 获取具体某个 cgroup 的具体绝对路径
func GetCgroupPath(subsystem string, cgroupPath string, autoCreate bool) (string, error) {
	// 获取某个 cgroup 的根路径
	cgroupRoot := FindCgroupMountPoint(subsystem)
	// os.Stat返回描述文件 f 的 FileInfo 类型值。如果出错，错误底层类型是 *PathError
	_, err := os.Stat(path.Join(cgroupRoot, cgroupPath))

	// 如果目录不存在就创建
	if err == nil || (autoCreate && os.IsNotExist(err)) {
		if os.IsNotExist(err) {
			if err := os.Mkdir(path.Join(cgroupRoot, "dockergsh", cgroupPath), 0755); err != nil {
				return "", fmt.Errorf("error create cgroup %v", err)
			}
		}
		return path.Join(cgroupRoot, cgroupPath), nil
	} else {
		return "", fmt.Errorf("cgroup path error %v", err)
	}
}
