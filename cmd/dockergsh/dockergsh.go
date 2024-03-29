package main

import (
	"github.com/Nevermore12321/dockergsh/cmd/dockergsh/root"
	"log"
	"os"
	"path/filepath"
)

func main() {
	// 获取当前执行的可执行文件的路径
	exePath, err := os.Executable()
	if err != nil {
		log.Fatalf("Can not find command by filename: %v", err)
	}

	// 使用文件路径获取文件名
	cmdName := filepath.Base(exePath)
	// 配置命令行 app
	root.Initial(cmdName, os.Stdin, os.Stdout, os.Stderr)

	// 执行命令行 app
	if err := root.RootCmd.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
