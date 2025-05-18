package user

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

type Group struct {
	Gid    int    // Gid 表示组的唯一标识符（组ID）
	Name   string // Name 表示组的名称
	Passwd string // Passwd 表示组的密码
	List   string // List 表示组成员的列表
}

// parseLine 解析每一行，将每一行按顺序解析为 Group 对象元素
func parseLine(line string, v ...interface{}) {
	if line == "" {
		return
	}

	// 解析每一行
	parts := strings.Split(line, ":")
	for i, p := range parts {
		if len(v) <= i { // 如果 group 文件每行的字段个数小于 v 的长度，则跳过
			break
		}
		switch e := v[i].(type) {
		case *int:
			*e, _ = strconv.Atoi(p)
		case *string:
			*e = p
		case *[]string:
			if p != "" {
				*e = strings.Split(p, ",")
			} else {
				*e = []string{}
			}
		default:
			panic("parseLine expects only pointers! argument " + strconv.Itoa(i) + " is not a pointer")
		}
	}

}

// ParseGroupFilter 从 r 中读取 group 信息
func ParseGroupFilter(r io.Reader, filter func(Group) bool) ([]Group, error) {
	if r == nil {
		return nil, fmt.Errorf("nil source for group-formatted data")
	}

	var (
		s   = bufio.NewScanner(r)
		out = make([]Group, 0)
	)

	// 读取文件中的每一行
	for s.Scan() {
		if err := s.Err(); err != nil {
			return nil, err
		}
		text := s.Text()
		if text == "" {
			continue
		}

		// see: man 5 group，格式为：
		//  group_name:password:GID:user_list
		//  Name:Pass:Gid:List
		//  root:x:0:root
		//  adm:x:4:root,adm,daemon
		item := Group{}
		// 解析每一个 Group 对象
		parseLine(text, &item.Name, &item.Passwd, &item.Gid, &item.List)
		if filter != nil && filter(item) {
			out = append(out, item)
		}
	}
	return out, nil
}

// ParseGroupFileFilter 通过 path 文件的 group 文件，解析文件到 Group 对象（通过 filter 过滤，返回符合条件的 Group 对象）。
func ParseGroupFileFilter(path string, filter func(Group) bool) ([]Group, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return ParseGroupFilter(file, filter)
}

func ParseGroupFile(path string) ([]Group, error) {
	return ParseGroupFileFilter(path, nil)
}
