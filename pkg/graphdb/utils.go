package graphdb

import (
	"path"
	"strings"
)

// splitPath 根据分隔符 / ，找到 path 中的 父parent，fullpath 形如 /xxx/xxx/xxx/xxx
func splitPath(fullPath string) (string, string) {
	var (
		parent string
		name   string
	)
	// 如果 path 中第一个不是 / 开头，在起始处添加 /
	if fullPath[0] != '/' {
		fullPath = "/" + fullPath
	}

	// parent 为 fullPath 的 dir，name 为 file name
	parent, name = path.Split(fullPath)
	length := len(parent)
	if parent[length-1] == '/' { // parent 最后一个字符是 /,则去掉
		parent = parent[:length-1]
	}
	return parent, name
}

// 将路径 p 按照分隔符 "/" 进行分隔
func split(p string) []string {
	return strings.Split(p, "/")
}

// 返回给定路径中的深度或数量
func pathDepth(p string) int {
	paths := split(p)
	if len(paths) == 2 && paths[1] == "" { // 如果第一个是空
		return 1
	}

	return len(paths)
}
