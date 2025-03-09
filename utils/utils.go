package utils

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"github.com/Nevermore12321/dockergsh/internal/utils"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

// DockergshInitPath 找出 dockergshinit 的路径
func DockergshInitPath(localCopy string) string {
	// 先判断可执行文件的路径，是否满足 hash
	selfPath := SelfPath()
	if isValidDockerInitPath(selfPath, selfPath) {
		return selfPath
	}

	// 否则，从下列列表中，寻找一个目录作为 init
	var possiblePaths = []string{
		localCopy,
		utils.INITPATH,
		filepath.Join(filepath.Dir(selfPath), "dockergshinit"),
		// FHS 3.0 Draft: "/usr/libexec includes internal binaries that are not intended to be executed directly by users or shell scripts. Applications may use a single subdirectory under /usr/libexec."
		"/usr/libexec/dockergsh/dockergshinit",
		"/usr/local/libexec/dockergsh/dockergshinit",

		// FHS 2.3: "/usr/lib includes object files, libraries, and internal binaries that are not intended to be executed directly by users or shell scripts."
		"/usr/lib/dockergsh/dockergshinit",
		"/usr/local/lib/dockergsh/dockergshinit",
	}

	for _, initPath := range possiblePaths {
		if initPath == "" {
			continue
		}
		path, err := exec.LookPath(initPath)
		if err == nil {
			path, err = filepath.Abs(initPath)
			if err != nil {
				panic(err)
			}
			if isValidDockerInitPath(path, selfPath) {
				return path
			}
		}
	}
	return ""
}

// SelfPath 找到当前可执行文件的绝对路径
func SelfPath() string {
	path, err := exec.LookPath(os.Args[0]) // 获取当前执行文件的路径
	if err != nil {
		if os.IsNotExist(err) {
			return ""
		}
		if execErr, ok := err.(*exec.Error); ok && os.IsNotExist(execErr.Err) {
			return ""
		}
		panic(err)
	}
	path, err = filepath.Abs(path)
	if err != nil {
		if os.IsNotExist(err) {
			return ""
		}
		panic(err)
	}
	return path
}

// Sum 返回数据的 SHA-1 校验和
func dockerInitSha1(target string) string {
	f, err := os.Open(target)
	if err != nil {
		return ""
	}
	defer f.Close()

	h := sha1.New()
	_, err = io.Copy(h, f)
	if err != nil {
		return ""
	}
	return hex.EncodeToString(h.Sum(nil))
}

// 校验 sha1 哈希值
func isValidDockerInitPath(target string, selfPath string) bool {
	if target == "" {
		return false
	}

	// 如果是通过 hack/make.sh
	if utils.IAMSTATIC {
		if selfPath == "" {
			return false
		}
		if target == selfPath {
			return true
		}
		targetFileInfo, err := os.Lstat(target)
		if err != nil {
			return false
		}
		selfPath, err := os.Lstat(selfPath)
		if err != nil {
			return false
		}
		return os.SameFile(targetFileInfo, selfPath) // 判断是否是同一个文件
	}
	return utils.INITSHA1 != "" && utils.INITSHA1 == dockerInitSha1(target)
}

// CopyFile 从 src 文件拷贝到 dst 文件，返回拷贝的文件字节数
func CopyFile(src, dst string) (int64, error) {
	if src == dst {
		return 0, nil
	}
	srcFile, err := os.Open(src) // 打开源文件
	if err != nil {
		return 0, err
	}
	defer srcFile.Close()

	// 删除目标文件，并新建
	if err := os.Remove(dst); err != nil && !os.IsNotExist(err) {
		return 0, err
	}
	dstFile, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer dstFile.Close()

	return io.Copy(dstFile, srcFile)
}

// GetLines 从 input 内容中去掉 commentMarker 注释内容
func GetLines(input []byte, commentMarker []byte) [][]byte {
	lines := bytes.Split(input, []byte("\n")) // 按照换行符分割
	var output [][]byte
	for _, currentLine := range lines { // 遍历每一行数据
		var commentIndex = bytes.Index(currentLine, commentMarker) // 寻找注释符号
		if commentIndex == -1 {                                    // 如果此行不是注释，添加到输出
			output = append(output, currentLine)
		} else { // 如果此行有注释，则去掉注释标识符后的内容
			output = append(output, currentLine[:commentIndex])
		}
	}
	return output
}

// CheckLocalDNS 会查看 /etc/resolv.conf，如果第一条生效的 那么nameserver是 127 的本地dns，则返回true
func CheckLocalDNS(resolvConf []byte) bool {
	// 循环读取 /etc/resolv.conf 文件中的每一行内容（去掉注释）
	for _, line := range GetLines(resolvConf, []byte("#")) {
		if !bytes.Contains(line, []byte("nameserver")) { // 如果内容没有 nameserver 字样，跳过
			continue
		}
		for _, ip := range [][]byte{ // 如果包括了 127.0.0.1 或者 127.0.1.1 的本地 dns 地址，返回 true
			[]byte("127.0.0.1"),
			[]byte("127.0.1.1"),
		} {
			if bytes.Contains(line, ip) {
				return true
			}
		}
		return false
	}
	return true
}

// TruncateID 截断容器 id 的前 12 位作为容器名称，有可能重复
func TruncateID(id string) string {
	shortLen := 12
	if len(id) < shortLen {
		shortLen = len(id)
	}

	return id[:shortLen]
}
