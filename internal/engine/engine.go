package engine

import (
	"fmt"
	"github.com/Nevermore12321/dockergsh/internal/utils"
	"io"
	"os"
	"sort"
	"sync"
)

// Handler job 的具体业务逻辑处理
type Handler func(*Job) Status

// 全局注册的所有 job
var globalHandlers map[string]Handler

func init() {
	globalHandlers = make(map[string]Handler)
}

/*
Engine扮演着Docker Container存储仓库的角色
并且通过Job的形式管理Docker运行中涉及的所有任务。
*/
type Engine struct {
	handlers   map[string]Handler // 表示每一个 job 对应的具体逻辑处理 handler
	catchall   Handler            // 错误处理
	id         string             // 当前 engine 的 id 号
	Stdout     io.Writer          // 标准输出
	Stderr     io.Writer          // 标准错误
	Stdin      io.Reader          // 标准输入
	Logging    bool               // 是否打开日志
	tasks      sync.WaitGroup     // 所有运行的 goroutine 数量
	lck        sync.RWMutex       // 读写互斥锁，用于 shutdown
	shutdown   bool               // 是否需要关闭逻辑处理
	onShutdown []func()           // 如果需要关闭的逻辑处理，在这里定义 handler
}

// Register engine 注册一个 Job，其实就是在 engine 的 handlers 中加一个 job
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
	eng.Register("commands", func(job *Job) Status {
		//	作用是通过Job来打印所有已经注册完毕的command名称，最终返回状态StatusOK
		for _, name := range eng.commands() {
			job.Printf("%s\n", name)
		}
		return StatusOk
	})

	// 将变量globalHandlers中定义完毕的所有Handler都复制到eng对象的handlers属性中
	for k, v := range globalHandlers {
		eng.handlers[k] = v
	}
	return eng
}
