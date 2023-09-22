//go:build darwin || dragonfly || freebsd || linux || netbsd || openbsd

package utils

import (
	"os"
	"path/filepath"
)

func TempDir(rootDir string) (string, error) {
	var tmpDir string
	if tmpDir = os.Getenv(ConfigTempdir); tmpDir == "" {
		tmpDir = filepath.Join(rootDir, "tmp")
	}
	err := os.MkdirAll(tmpDir, 0700)
	return tmpDir, err
}
