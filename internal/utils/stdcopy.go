package utils

import (
	"errors"
	log "github.com/sirupsen/logrus"
	"io"
)

const (
	StdWriterPrefixLen = 8 // header 总长度
	StdWriterFdIndex   = 0 // 第一个字节标识输出类型
)

var ErrInvalidStdHeader = errors.New("unrecognized input header")

// StdCopy 是io.copy的修改版本
// StdCopy 用于处理 多路复用（multiplexed）输出流，将写入`dstout'和`dsterr'。
func StdCopy(dstout, dsterr io.Writer, src io.Reader) (int64, error) {
	var (
		written int64
		err     error
		buf     = make([]byte, 32*1024+StdWriterPrefixLen+1) // 初始 buffer，用于读入 header + payload
		bufLen  = len(buf)                                   // 当前 buffer 长度（避免每次计算）
		nr, nw  int                                          // nr 当前 buffer 中已读的数据字节数，nw 当前写出的字节数
		er, ew  error                                        // er: 读入时产生的错误，ew: 写出时产生的错误
		out     io.Writer
	)

	for {
		// 首先读取 Header（8 字节）
		for nr < StdWriterPrefixLen {
			// nr2 是新读取的字节数
			var nr2 int
			nr2, er = src.Read(buf[nr:]) // 从 src 中先读取
			nr += nr2                    // nr 后移
			if er == io.EOF {
				if nr < StdWriterPrefixLen { // 如果到了 EOF（结束），但 header 不足 8 字节，说明数据损坏
					log.Debugf("Corrupted prefix: %v", buf[:nr])
					return written, nil
				}
				break
			}
			if er != nil {
				log.Debugf("Error reading header: %s", er)
				return 0, err
			}
		}
		// buf[0] 是输出类型标识
		// 1：stdout
		// 2：stderr
		// 0：也视为 stdout（兼容模式）
		switch buf[StdWriterFdIndex] {
		case 0:
			fallthrough
		case 1:
			out = dstout
		case 2:
			out = dsterr
		default:
			log.Debugf("Error selecting output fd: (%d)", buf[StdWriterFdIndex])
			return 0, ErrInvalidStdHeader
		}
		// todo
	}
}
