package engine

import (
	"bytes"
	"fmt"
	"strings"
	"time"
)

/*
 作业 Job 是 dockergsh engine 中的基本工作单元。dockergsh 能做的一切最终都应该作为一项工作公开。
 	例如：在容器中执行一个进程，创建一个新容器， 从网上下载镜像、提供 http api 等。
 	或者例如：dockergsh run ，其实就是一个 run 的 job 在运行

 作业 Job API是根据 unix 流程设计的：
	1. 名称
	2. 参数
	3. 环境变量
	4. 输入、输出和错误的标准流，
	5. 退出状态，可以指示成功（0）或错误（任何其他）。
	6. 一个变化是作业Job将其状态报告为字符串。字符串“0”表示成功，任何其他字符串表示错误。
*/

type Job struct {
	Eng     *Engine   // dockergsh 引擎 Engine 对象
	Name    string    // Job 作业名称
	Args    []string  // Job 作业运行参数
	env     *Env      // Job 作业环境变量
	Stdout  *Output   // Job 作业输出标准流
	Stderr  *Output   // Job 作业错误标准流
	Stdin   *Input    // Job 作业输入标准流
	handler Handler   // Job 作业的处理函数
	status  Status    // Job 作业的退出状态
	end     time.Time // Job 作业的完成时间
}

// Status 退出状态，其实就是 int 整型
type Status int

// 自定义一些状态
const (
	StatusOk       Status = 0   // 成功状态
	StatusErr      Status = 1   // 失败状态
	StatusNotFound Status = 127 // Job command not found 错误状态
)

// CallString Job 作业调用的打印格式，格式为 JOB_NAME(args1, args2, ....)
func (job *Job) CallString() string {
	return fmt.Sprintf("%s(%s)", job.Name, strings.Join(job.Args, ", "))
}

// StatusString  Job 作业运行完成后的状态打印格式，格式为 = STATUS_STR (STATUS_INT)
func (job *Job) StatusString() string {
	// 如果作业 job 没有完成，返回 ""
	if job.end.IsZero() {
		return ""
	}

	// 状态信息
	var okErr string
	if job.status == StatusOk {
		okErr = "OK"
	} else {
		okErr = "ERROR"
	}

	return fmt.Sprintf(" = %s (%d)", okErr, job.status)
}

// CallString Job 作业的打印格式，格式为 ENG_ID.JOB_NAME(args1, args2, ....) = STATUS_STR (STATUS_INT)
func (job *Job) String() string {
	return fmt.Sprintf("%s.%s%s", job.Eng, job.CallString(), job.StatusString)
}

// Run 执行 Job 作业并阻塞，直到作业完成。
// 如果作业返回失败状态，则返回错误
func (job *Job) Run() error {
	// 判断 engine 是否已经关闭
	if job.Eng.IsShutdown() {
		return fmt.Errorf("engine is shutdown")
	}

	//ServeApi 是后台运行的守护进程，因此 Job 会阻塞运行
	if job.Name != "serveapi" {
		// 修改 engine 对象的内容，都需要线程安全，加锁
		job.Eng.lck.Lock()
		job.Eng.tasks.Add(1)
		job.Eng.lck.Unlock()
		defer job.Eng.tasks.Done()
	}

	// 如果 Job 作业已经有了完成时间，表示作业已经完成
	if !job.end.IsZero() {
		return fmt.Errorf("%s: job has already completed", job)
	}

	// 在作业 Job 运行前，打印日志
	job.Eng.Logf("+job %s", job.CallString())
	// 在作业 Job 运行结束，打印日志
	defer func() {
		job.Eng.Logf("-job %s%s", job.CallString(), job.StatusString())
	}()

	// 错误信息
	var errMessage = bytes.NewBuffer(nil)
	job.Stderr.Add(errMessage)

	// 如果有 handler，执行 job 任务
	if job.handler == nil {
		job.Errorf("%s: command not found", job.Name)
		job.status = StatusNotFound
	} else {
		// 执行 job 任务，并获取 返回状态
		job.status = job.handler(job)
		// 完成时间
		job.end = time.Now()
	}

	// 等待所有后台任务完成
	if err := job.Stdout.Close(); err != nil {
		return err
	}
	if err := job.Stderr.Close(); err != nil {
		return err
	}
	if err := job.Stdin.Close(); err != nil {
		return err
	}

	// 如果返回状态不为0，即返回错误
	if job.status != 0 {
		return fmt.Errorf("%s", Tail(errMessage))
	}
	return nil
}

/* **************** ENV 环境变量相关 **************** */

// Env 返回当前 job 的环境变量信息
func (job *Job) Env() *Env {
	return job.env
}

// SetEnvBool 设置 job 的环境变量(bool类型)
func (job *Job) SetEnvBool(key string, value bool) {
	job.env.SetBool(key, value)
}

// GetEnvBool 获取 job 的环境变量(bool类型)
func (job *Job) GetEnvBool(key string) bool {
	return job.env.GetBool(key)
}

// SetEnv 设置 job 的环境变量(string类型)
func (job *Job) SetEnv(key, value string) {
	job.env.Set(key, value)
}

// GetEnv 获取 job 的环境变量(string类型)
func (job *Job) GetEnv(key string) string {
	return job.env.Get(key)
}

// Printf Job 作业的 日志打印
func (job *Job) Printf(format string, args ...interface{}) (n int, err error) {
	return fmt.Fprintf(job.Stdout, format, args...)
}

// Errorf Job 作业的 错误日志
func (job *Job) Errorf(format string, args ...any) Status {
	format = strings.TrimRight(format, "\n")
	fmt.Fprintf(job.Stderr, format+"\n", args...)
	return StatusErr
}
