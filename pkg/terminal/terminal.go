package terminal

import (
	"errors"
	"os"
	"os/signal"
	"syscall"
	"unsafe"
)

// 保存终端的 termios 设置，用于后续恢复
type State struct {
	termios syscall.Termios
}

const (
	getTermios = syscall.TCGETS
	setTermios = syscall.TCSETS
)

var (
	ErrInvalidState = errors.New("invalid terminal state")
)

// 与 C 语言中 struct winsize 对应，包含终端窗口的行列信息。x 和 y 未使用（占位保留）。
type Winsize struct {
	Height uint16
	Width  uint16
	x      uint16
	y      uint16
}

func GetWinsize(fd uintptr) (*Winsize, error) {
	ws := &Winsize{}
	// 用于获取终端窗口大小（通常用于 pty 的 resize 同步）。
	_, _, err := syscall.Syscall(syscall.SYS_IOCTL, fd, uintptr(syscall.TIOCGWINSZ), uintptr(unsafe.Pointer(ws)))
	if err != 0 {
		return ws, err
	}
	return ws, nil
}

func SetWinsize(fd uintptr, ws *Winsize) error {
	_, _, err := syscall.Syscall(syscall.SYS_IOCTL, fd, uintptr(syscall.TIOCSWINSZ), uintptr(unsafe.Pointer(ws)))
	if err == 0 {
		return nil
	}
	return err
}

// todo 如果给定的文件描述符是终端，则 IsTerminal 返回 true。
func IsTerminal(fd uintptr) bool {
	var termios syscall.Termios
	_, _, err := syscall.Syscall(syscall.SYS_IOCTL, fd, uintptr(getTermios), uintptr(unsafe.Pointer(&termios)))
	return err == 0
}

func SaveState(fd uintptr) (*State, error) {
	var oldState State
	if _, _, err := syscall.Syscall(syscall.SYS_IOCTL, fd, getTermios, uintptr(unsafe.Pointer(&oldState.termios))); err != 0 {
		return nil, err
	}

	return &oldState, nil
}

func RestoreTerminal(fd uintptr, state *State) error {
	if state == nil {
		return ErrInvalidState
	}
	_, _, err := syscall.Syscall(syscall.SYS_IOCTL, fd, uintptr(setTermios), uintptr(unsafe.Pointer(&state.termios)))
	if err != 0 {
		return err
	}
	return nil
}

// DisableEcho 终端关闭 echo 回显,常用于输入密码
func DisableEcho(fd uintptr, state *State) error {
	newState := state.termios
	// 关闭 echo 状态
	newState.Lflag &^= syscall.ECHO

	if _, _, err := syscall.Syscall(syscall.SYS_IOCTL, fd, setTermios, uintptr(unsafe.Pointer(&newState))); err != 0 {
		return err
	}
	handleInterrupt(fd, state)
	return nil
}

// 注册 Ctrl+C 信号捕捉，确保程序退出时恢复终端状态。
func handleInterrupt(fd uintptr, state *State) {
	signalChan := make(chan os.Signal, 1)
	// 注册 Ctrl+C 信号
	signal.Notify(signalChan, os.Interrupt)

	go func() {
		// 捕捉到信号，恢复 终端状态
		_ = <-signalChan
		_ = RestoreTerminal(fd, state)
		os.Exit(0)
	}()

}
