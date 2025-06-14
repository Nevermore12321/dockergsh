package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"time"
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

func (progress *JSONProgress) String() string {
	return ""
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
// Pulling from library/ubuntu
// fdd5d7827f33: Downloading [================>                  ]  12.34MB/30MB
func DisplayJSONMessagesStream(in io.Reader, out io.Writer, terminalFd uintptr, isTerminal bool) error {
	var (
		dsc  = json.NewDecoder(in)
		ids  = make(map[string]int) // 表示 id 层在对应显示的第几行
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

		// 每一层的输出信息进度
		if jsonMessage.ID != "" && (jsonMessage.Progress != nil || jsonMessage.ProgressMessage != "") {
			line, ok := ids[jsonMessage.ID]
			if !ok { // 如果是新的 id 层，另起一行显示
				line := len(ids)
				ids[jsonMessage.ID] = line
				fmt.Fprintf(out, "\n")
				diff = 0
			} else {
				diff = len(ids) - line
			}
			if jsonMessage.ID != "" && isTerminal {
				// 利用 ANSI 控制符 ESC[{diff}A（即“光标上移 diff 行”）回到对应行
				fmt.Fprintf(out, "%c[%dA", 27, diff)
			}
		}

		err := jsonMessage.Display(out, isTerminal)
		if jsonMessage.ID != "" && isTerminal {
			// 因为上面曾“向上移动光标”，这一步要“下移回来”
			fmt.Fprintf(out, "%c[%dB", 27, diff)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

// Display 格式化输出
func (jsonMessage JSONMessage) Display(out io.Writer, isTerminal bool) error {
	if jsonMessage.Error != nil {
		if jsonMessage.Error.Code == 401 {
			return fmt.Errorf("authentication is required")
		}
		return jsonMessage.Error
	}

	// 如果是终端进度条，为 \r（不换行，只回到行首）；否则为空或 \n
	var endl string

	// 如果是终端，并且是进度条消息（无 Stream），清空当前行（为了“刷新”旧的进度内容）。
	if isTerminal && jsonMessage.Stream == "" && jsonMessage.Progress != nil {
		// <ESC>[2K\r 清空当前行（2K）并回到行首
		fmt.Fprintf(out, "%c[2K\r", 27)
		endl = "\r"
	} else { // 否则，不显示进度条
		fmt.Fprintf(out, "%s ", time.Unix(jsonMessage.Time, 0).Format(time.RFC3339Nano))
	}

	// 打印时间辍
	if jsonMessage.Time != 0 {
		fmt.Fprintf(out, "%s ", time.Unix(jsonMessage.Time, 0).Format(time.RFC3339Nano))
	}

	// 打印 ID
	if jsonMessage.ID != "" {
		fmt.Fprintf(out, "%s:", jsonMessage.ID)
	}

	// 打印 来源 From（例如拉取哪个镜像层）
	if jsonMessage.From != "" {
		fmt.Fprintf(out, "(from %s) ", jsonMessage.From)
	}

	// 打印进度条
	if jsonMessage.Progress != nil {
		fmt.Fprintf(out, "%s %s%s", jsonMessage.Status, jsonMessage.Progress.String(), endl)
	} else if jsonMessage.Stream != "" { //  流式输出（如构建日志）
		fmt.Fprintf(out, "%s%s", jsonMessage.Stream, endl)
	} else { // 普通状态消息
		fmt.Fprintf(out, "%s%s\n", jsonMessage.Status, endl)
	}
	return nil
}
