package version

import (
	"fmt"
	"io"
	"os"
)

func Package() string {
	return mainPkg
}

func Version() string {
	return version
}

// FprintVersion 使用 Fprintln 向特定的 描述符写入 version 信息
func FprintVersion(w io.Writer) {
	fmt.Fprintln(w, os.Args[0], Package(), Version())
}

// PrintVersion 向控制台输出 registry 版本信息
func PrintVersion() {
	FprintVersion(os.Stdout)
}
