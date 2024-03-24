package engine

import (
	"fmt"
	"github.com/Nevermore12321/dockergsh/internal/utils"
	log "github.com/sirupsen/logrus"
	"io"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
)

// Handler job 的具体业务逻辑处理的函数声明
type Handler func(*Job) Status

// 全局注册的所有 job
var globalHandlers map[string]Handler

func init() {
	globalHandlers = make(map[string]Handler)
}

/*
Engine 扮演着Docker Container存储仓库的角色
并且通过Job的形式管理Docker运行中涉及的所有任务。
即，docker server 收到命令，通过在 engine 中创建 job 执行具体的人物
*/
type Engine struct {
	handlers   map[string]Handler // 表示每一个 job 对应的具体逻辑处理 handler
	catchall   Handler            // 错误处理
	id         string             // 当前 engine 的 id 号
	Stdout     io.Writer          // 标准输出
	Stderr     io.Writer          // 标准错误
	Stdin      io.Reader          // 标准输入
	Logging    bool               // 是否打开日志
	tasks      sync.WaitGroup     // 所有运行 tasks 的 goroutine 数量
	lck        sync.RWMutex       // 读写互斥锁，用于 shutdown
	shutdown   bool               // 是否已经执行了关闭
	onShutdown []func()           // 如果需要关闭的逻辑处理，在这里定义 handler
	// todo hack
}

// Register engine 注册一个 Job，其实就是在 engine 的 handlers 中加一个 job
// 向eng对象注册处理方法，并不代表处理方法的值函数会被立即调用执行
func (eng *Engine) Register(name string, handler Handler) error {
	_, exists := eng.handlers[name]
	if exists {
		return fmt.Errorf("Can't register a existed job %s", name)
	}
	eng.handlers[name] = handler
	return nil
}

// 返回当前注册当 Engine 中的命令列表
func (eng *Engine) commands() []string {
	names := make([]string, 0, len(eng.handlers))
	for name := range eng.handlers {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// Shutdown engine 关闭：
// Daemon不再接受任何新的Job
// Daemon等待所有存活的Job执行完毕
// Daemon调用所有shutdown的处理方法
// 在15秒时间内，若所有的handler执行完毕，则Shutdown（）函数返回，否则强制返回。
// 以先发生者为准。
func (eng *Engine) Shutdown() {
	eng.lck.Lock()    // 加锁
	if eng.shutdown { // 如果已经执行了关闭逻辑处理
		eng.lck.Unlock()
		return
	}
	eng.shutdown = true // 这里要加锁的原因是，有可能一个 goroutine 已经在关闭了，另一个关闭又开始了
	// 不需要用锁来保护其余部分，以允许其他调用立即失败并“关闭”，而不是挂起 15 秒。
	// 这需要所有并发调用检查是否关闭，否则可能会导致竞争
	eng.lck.Unlock() // 解锁

	// 开始关闭处理
	// 1. 等待所有任务 tasks 关闭，等待 5s
	tasksDone := make(chan struct{})
	go func() {
		eng.tasks.Wait() // 等待所有 tasks 结束，即 waitGroup 没有 goroutine 运行
		close(tasksDone) // 关闭 tasksDone channel，表示已经全部关闭
	}()
	select {
	case <-time.After(time.Second * 5):
	case <-tasksDone:
	}

	// 2. 调用 shutdown handler
	var wg sync.WaitGroup
	for _, handler := range eng.onShutdown {
		wg.Add(1)
		go func(f func()) {
			defer wg.Done()
			f()
		}(handler)
	}

	// 3. 等待 10s
	Done := make(chan struct{})
	go func() {
		wg.Wait()   // 等待所有 shutdown handler 结束
		close(Done) // 关闭 Done channel，表示已经全部关闭
	}()
	select {
	case <-time.After(time.Second * 10):
	case <-tasksDone:
	}
	return
}

// OnShutdown 向 engine 中注册一个 shutdown 时的回调函数
func (eng *Engine) OnShutdown(h func()) {
	eng.lck.Lock()
	// 添加到 engine 中的 onShutdown 列表
	eng.onShutdown = append(eng.onShutdown, h)
	eng.lck.Unlock()
}

// IsShutdown 判断 dockergsh 引擎 engine 是否已经被关闭
func (eng *Engine) IsShutdown() bool {
	// 加读锁
	eng.lck.RLock()
	defer eng.lck.RUnlock()
	// 返回 engine 对象的 shutdown 字段，表示已经关闭
	return eng.shutdown
}

// Job 创建并初始化 Job 作业对象
func (eng *Engine) Job(name string, args ...string) *Job {
	job := &Job{
		Eng:    eng,
		Name:   name,
		Args:   args,
		Stdin:  NewInput(),
		Stdout: NewOutput(),
		Stderr: NewOutput(),
		env:    &Env{},
	}
	// 如果engine 开启了日志，添加 job 的日志输出
	// todo
	if eng.Logging {
		job.Stderr.Add()
	}

	// 在 engine 的 hanlers 任务作业Job列表中寻找当前 job
	if hanler, exits := eng.handlers[name]; exits {
		// 添加 handler 到 job 对象中
		job.handler = hanler
	} else if eng.catchall != nil && name != "" { // 名称不能为空，
		// 如果不存在该 job 的 handler
		job.handler = eng.catchall
	}
	return job
}

// engine 打印输出格式，取 id 的前8位
func (eng *Engine) String() string {
	return fmt.Sprintf("%s", eng.id[:8])
}

// Logf engine 的日志打印方法
func (eng *Engine) Logf(format string, args ...interface{}) (int, error) {
	// 如果日志没有打开
	if !eng.Logging {
		return 0, nil
	}

	// 打印时添加 [engine_id] 的前缀
	prefixedFormat := fmt.Sprintf("[%s] %s\n", eng, strings.TrimRight(format, "\n"))
	return fmt.Fprintf(eng.Stdout, prefixedFormat, args...)
}

// New 创建 Engine 实例对象
func New() *Engine {
	// 创建一个Engine结构体实例eng，并初始化部分属性
	eng := &Engine{
		handlers: make(map[string]Handler),
		id:       utils.RandomString(),
		Stdout:   os.Stdout,
		Stderr:   os.Stderr,
		Stdin:    os.Stdin,
		Logging:  true,
	}

	// 向eng对象注册名为commands的Handler
	err := eng.Register("commands", func(job *Job) Status {
		//	作用是通过Job来打印所有已经注册完毕的command名称，最终返回状态StatusOK
		for _, name := range eng.commands() {
			job.Printf("%s\n", name)
		}
		return StatusOk
	})
	if err != nil {
		log.Fatal(err)
	}

	// 将变量globalHandlers中定义完毕的所有Handler都复制到eng对象的handlers属性中
	for k, v := range globalHandlers {
		eng.handlers[k] = v
	}
	return eng
}
