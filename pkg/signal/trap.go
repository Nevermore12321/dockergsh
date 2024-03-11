package signal

import (
	"github.com/Nevermore12321/dockergsh/internal/utils"
	log "github.com/sirupsen/logrus"
	"os"
	osSignal "os/signal"
	"sync/atomic"
	"syscall"
)

// Trap 信号捕捉并处理函数
func Trap(cleanup func()) {
	// 1. 创建并设置一个channel，用于发送信号通知
	// channel 的大小为 1
	c := make(chan os.Signal, 1)

	// 2.1 定义signals数组变量，初始值为os.SIGINT，os.SIGTERM
	signals := []os.Signal{os.Interrupt, syscall.SIGTERM}
	// 2.2 若环境变量DEBUG为空，则添加os.SIGQUIT至signals数组
	if os.Getenv(utils.DockergshDebug) == "" {
		signals = append(signals, syscall.SIGQUIT)
	}

	// 3. 通过标准库 signal.Notify(c，signals...) 中 Notify 函数来实现将接收到的signal信号传递给c
	osSignal.Notify(c, signals...)

	// 4. 创建一个goroutine来处理具体的signal信号
	go func() {
		// 尝试次数
		interruptCount := uint32(0)
		// 遍历获取 channal c 中的信号，获取到信号后，创建新的 goroutine 处理
		for sig := range c {
			go func(sig os.Signal) {
				log.Printf("Received signal '%v', starting shutdown of docker...", sig)
				// 判断捕获的信号类型
				switch sig {
				case os.Interrupt, syscall.SIGTERM: // 当信号类型为os.Interrupt或者syscall.SIGTERM时
					// 用户连续发送的 中断信号 少于 3 次，优雅关闭
					if atomic.LoadUint32(&interruptCount) < 3 {
						atomic.AddUint32(&interruptCount, 1)
						// 如果是第一次收到 中断信号, 执行参数的 cleanup 函数
						if atomic.LoadUint32(&interruptCount) == 1 {
							// 善后工作有：
							// Daemon不再接受任何新的Job
							// Daemon等待所有存活的Job执行完毕
							// Daemon调用所有shutdown的处理方法
							// 在15秒时间内，若所有的handler执行完毕，则Shutdown（）函数返回，否则强制返回。
							cleanup()
							os.Exit(0)
						} else { // 否则，什么也不做
							return
						}
					} else { // 用户连续发送的 中断信号 大于 3 次，立即关闭
						log.Printf("Force shutdown of dockergsh, interrupting cleanup")
					}
				case syscall.SIGQUIT:
					// SIGQUIT 信号，什么也不做
				}
				// 此 goroutine 退出
				os.Exit(128 + int(sig.(syscall.Signal)))
			}(sig)
		}
	}()
}
