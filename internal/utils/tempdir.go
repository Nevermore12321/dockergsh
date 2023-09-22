//go:build !darwin && !dragonfly && !freebsd && !linux && !netbsd && !openbsd

package utils

import "os"

// TempDir 创建用于临时文件的默认目录。
func TempDir(rootDir string) (string, error) {
	return os.TempDir(), nil
}
