package user

import (
	"bufio"
	"fmt"
	"io"
)

type Group struct {
	Gid    int    // Gid 表示组的唯一标识符（组ID）
	Name   string // Name 表示组的名称
	Passwd string // Passwd 表示组的密码
	List   string // List 表示组成员的列表
}

// 从 r 中读取 group 信息
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
		parseLine(text, &item)
		if filter != nil && filter(item) {
			out = append(out, item)
		}
	}
	return out, nil
}
