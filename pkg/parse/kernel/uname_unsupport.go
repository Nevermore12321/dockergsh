//go:build !linux
// +build !linux

package kernel

import (
	"errors"
	"syscall"
)

func uname() (*syscall.Utsname, error) {
	return nil, errors.New("Kernel version detection is available only on linux")
}
