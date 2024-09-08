package daemongsh

import (
	"sync"
	"time"
)

type State struct {
	sync.RWMutex               // 读写锁，多读单写
	Running      bool          // 运行中
	Paused       bool          // 暂停
	Restarting   bool          // 正在重启
	Pid          int           // 当前容器主进程的进程 pid
	ExitCode     int           // 容器退出状态码
	StartedAt    time.Time     // 容器启动时间
	FinishedAt   time.Time     // 容器结束时间
	waitChan     chan struct{} // 等待容器启动成功的进程间通信的 chan
}

// NewState 新建一个空的 State 实例
func NewState() *State {
	return &State{
		waitChan: make(chan struct{}),
	}
}
