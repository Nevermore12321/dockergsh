package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"math/rand"
	"os"
	"time"
)

/*
创建一个管道
返回：
- 只读管道 - *os.File
- 只写管道 - *os.File
- err - error
*/
func NewPipe() (*os.File, *os.File, error) {
	reader, writer, err := os.Pipe()
	if err != nil {
		return nil, nil, err
	}
	return reader, writer, nil
}

/*
生成 32 位的随机 id
*/
func NewId() string {
	letterBytes := "1234567890abcdefghigklmnopqrstuvwxyz"
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, 16)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

/*
将 id 进行 base32 编码
*/
func EncodeSha256(b []byte) string {
	bytes := sha256.Sum256(b)
	hashCode := hex.EncodeToString(bytes[:])
	return hashCode
}

/*
判断文件夹是否存在
*/
func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
