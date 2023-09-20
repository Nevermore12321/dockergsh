package kernel

/*
	在 linux 机器上执行 uname 获取 kernel 的版本信息
*/

// linux 操作系统上执行 uname 获取结果

import "syscall"

func uname() (*syscall.Utsname, error) {
	utsname := &syscall.Utsname{}

	if err := syscall.uname(utsname); err != nil {
		return nil, err
	}
	return utsname, nil
}
