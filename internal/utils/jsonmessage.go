package utils

import (
	"encoding/json"
	"io"
)

// JSONError 返回响应的错误信息
type JSONError struct {
	Code    int    `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

func (err *JSONError) Error() string {
	return err.Message
}

type JSONProgress struct {
	terminalFd uintptr
	Current    int `json:"current,omitempty"`
	Total      int `json:"total,omitempty"`
	Start      int `json:"start,omitempty"`
}

// JSONMessage 返回响应 Json 格式
type JSONMessage struct {
	Stream          string        `json:"stream,omitempty"`
	Status          string        `json:"status,omitempty"`
	Progress        *JSONProgress `json:"progress,omitempty"`
	ProgressMessage string        `json:"progressMessage,omitempty"`
	ID              string        `json:"id,omitempty"`
	From            string        `json:"from,omitempty"`
	Time            int64         `json:"time,omitempty"`
	Error           *JSONError    `json:"error,omitempty"`
	ErrorMessage    string        `json:"errorMessage,omitempty"`
}

// DisplayJSONMessagesStream 从 in 向 out 输出 json 格式的日志
func DisplayJSONMessagesStream(in io.Reader, out io.Writer, terminalFd uintptr, isTerminal bool) error {
	var (
		dsc  = json.NewDecoder(in)
		ids  = make(map[string]int)
		diff = 0
	)

	// 不断从 in 中读取
	for {
		var jsonMessage JSONMessage
		if err := dsc.Decode(&jsonMessage); err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		if jsonMessage.Progress != nil {
			jsonMessage.Progress.terminalFd = terminalFd
		}

		// todo
	}
	return nil
}
