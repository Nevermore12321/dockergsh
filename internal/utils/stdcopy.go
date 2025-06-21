package utils

import (
	"encoding/binary"
	"errors"
	log "github.com/sirupsen/logrus"
	"io"
)

const (
	StdWriterPrefixLen = 8 // header 总长度
	StdWriterFdIndex   = 0 // 第一个字节标识输出类型
	StdWriterSizeIndex = 4 // header 第 5~8 字节：4 字节大端整数，payload 长度
)

var ErrInvalidStdHeader = errors.New("unrecognized input header")

// StdCopy 是io.copy的修改版本
// StdCopy 用于处理 多路复用（multiplexed）输出流，将写入`dstout'和`dsterr'。
/*
+------+----+----+----+----+----+----+----+----+------------------+
| Fd   |          Frame Size (uint32 big-endian)                | Payload (N bytes) |
+------+----+----+----+----+----+----+----+----+------------------+
1 byte     4 bytes (total frame size, excluding header)          N bytes

Fd：0 表示 stdin（不常见），1 表示 stdout，2 表示 stderr。
Size：payload 的长度。
Payload：数据内容。
*/
// 例如 [1][0 0 0 5]['H' 'e' 'l' 'l' 'o']
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
		// buf[0] 决定是 stdout 还是 stderr
		// buf[4:8] 是 payload 长度（4 字节 big-endian）
		frameSize := int(binary.BigEndian.Uint32(buf[StdWriterSizeIndex : StdWriterSizeIndex+4]))
		log.Debugf("frameSize: %v", frameSize)

		// 检查缓冲区是否足够大以读取框架，否则扩容
		if frameSize+StdWriterPrefixLen > bufLen {
			log.Debugf("Extending buffer cap by %d (was %d)", frameSize+StdWriterPrefixLen-bufLen+1, len(buf))
			buf = append(buf, make([]byte, frameSize+StdWriterPrefixLen-bufLen)...)
			bufLen = len(buf)
		}

		// 读取完整的 frame（header + payload）
		for nr < frameSize+StdWriterPrefixLen {
			var nr2 int
			nr2, er = src.Read(buf[nr:])
			nr += nr2
			if er == io.EOF {
				if nr < frameSize+StdWriterPrefixLen {
					log.Debugf("Corrupted frame: %v", buf[StdWriterPrefixLen:nr])
					return written, nil
				}
				break
			}
			if er != nil {
				log.Debugf("Error reading frame: %s", er)
				return 0, err
			}
		}

		// 写入 payload 到对应的输出（stdout/stderr）
		nw, ew = out.Write(buf[StdWriterPrefixLen : frameSize+StdWriterPrefixLen])
		if ew != nil {
			log.Debugf("Error writing frame: %s", ew)
			return 0, ew
		}

		if nw != frameSize {
			log.Debugf("Error Short Write: (%d on %d)", nw, frameSize)
			return 0, io.ErrShortWrite
		}
		written += int64(nw)

		// 移动 buffer（类似 ring buffer）
		copy(buf, buf[frameSize+StdWriterPrefixLen:])
		// 减去这次读走的部分，更新 nr
		nr -= frameSize + StdWriterPrefixLen
	}
}
