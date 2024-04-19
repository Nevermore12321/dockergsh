package terminal

// todo 如果给定的文件描述符是终端，则 IsTerminal 返回 true。
func IsTerminal(fd uintptr) bool {
	return true
}
