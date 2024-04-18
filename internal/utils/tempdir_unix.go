//go:build darwin || dragonfly || freebsd || linux || netbsd || openbsd

package utils

import (
	"os"
	"path/filepath"
)

// TempDir 创建临时目录
// 优先使用 DOCKERGSH_TMPDIR 环境变量的 临时目录，如果没有，使用 rootDir
func TempDir(rootDir string) (string, error) {
	var tmpDir string
	if tmpDir = os.Getenv(ConfigTempdir); tmpDir == "" {
		tmpDir = filepath.Join(rootDir, "tmp")
	}
	err := os.MkdirAll(tmpDir, 0700)
	return tmpDir, err
}
