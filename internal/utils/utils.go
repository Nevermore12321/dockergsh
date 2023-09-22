package utils

import (
	"fmt"
	"os"
	"path/filepath"
)

// ReadSymlinkedDirectory 返回符号链接的目标目录。 符号链接的目标可能不是文件。
func ReadSymlinkedDirectory(path string) (string, error) {
	var (
		err      error
		realPath string
	)
	// 绝对路径
	realPath, err = filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("unable to get absolute path for %s: %s", path, err)
	}

	// 通过文件的软连接，找到源目标文件
	realPath, err = filepath.EvalSymlinks(realPath)
	if err != nil {
		return "", fmt.Errorf("failed to canonicalise path for %s: %s", path, err)
	}

	// 如果文件不存在，或者无法获取文件状态信息
	realPathInfo, err := os.Stat(realPath)
	if err != nil {
		return "", fmt.Errorf("failed to stat target '%s' of '%s': %s", realPath, path, err)
	}

	// 如果目标文件不是目录
	if !realPathInfo.Mode().IsDir() {
		return "", fmt.Errorf("canonical path points to a file '%s'", realPath)
	}
	return realPath, err
}
