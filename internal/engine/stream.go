package engine

import (
	"bytes"
	"io"
	"sync"
)

// Output 输出流
// 输出流可以有多个目标输出，但是同一时间只能一个 task 写入
type Output struct {
	sync.Mutex                // 互斥锁
	dests      []io.Writer    // 目标输出
	tasks      sync.WaitGroup // 目前有多少正在运行的任务
	used       bool           // 正在使用
}

func NewOutput() *Output {
	return &Output{}
}

// Add 添加将新目标流附加到输出。
// 随后写入输出的任何数据都将写入新的目标流。该方法是线程安全的。
func (output *Output) Add(dst io.Writer) {
	output.Lock()
	defer output.Unlock()
	output.dests = append(output.dests, dst) // 添加输出流
}

// Close 取消注册所有目标流并等待所有后台输出流都关闭
// 如果每个目标存在，则调用其 Close 方法。
func (output *Output) Close() error {
	output.Lock()
	defer output.Unlock()

	var firstErr error
	// 遍历所有输出流，执行 close
	for _, dst := range output.dests {
		if closer, ok := dst.(io.Closer); ok {
			err := closer.Close()
			if err != nil && firstErr == nil {
				firstErr = err
			}
		}
	}

	output.tasks.Wait() // 等待所有任务结束
	return firstErr
}

func (output *Output) Write(content []byte) (int, error) {
	output.Lock()
	defer output.Unlock()

	output.used = true // 正在写入
	var firstErr error
	// 向所有的输出流写入
	for _, dst := range output.dests {
		_, err := dst.Write(content)
		if err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return len(content), firstErr
}

// Input 输入流
type Input struct {
	sync.Mutex           // 互斥锁
	src        io.Reader // 输入源，只能有一个
}

func NewInput() *Input {
	return &Input{}
}

// Close 关闭输入流，不是线程安全
func (input *Input) Close() error {
	if input.src != nil {
		if closer, ok := input.src.(io.Closer); ok {
			return closer.Close()
		}
	}
	return nil
}

// Tail 返回 buffer 中的最后 n 行， n <= 0，则返回一个空字符串
func Tail(buffer *bytes.Buffer, n int) string {
	if n < 0 {
		return ""
	}

	bufferBytes := buffer.Bytes() // buffer 中的字节
	// 如果 buffer 中有数据，并且最后一个字节是 \n 符
	if len(bufferBytes) > 0 && bufferBytes[len(bufferBytes)-1] == '\n' {
		bufferBytes = bufferBytes[:len(bufferBytes)-1] // 去掉换行符，不计算最后一行的换行符
	}

	// 从后往前遍历 buffer，找到最后 n 个换行符
	for i := len(bufferBytes) - 2; i >= 0; i-- {
		if bufferBytes[i] == '\n' {
			n--
			if n == 0 { // 找到后，直接返回其后的字符串
				return string(bufferBytes[i+1:])
			}
		}
	}

	// 如果没有找到后n个，那么就返回所有
	return string(bufferBytes)
}
