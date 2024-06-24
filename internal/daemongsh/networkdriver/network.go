package networkdriver

import "errors"

// error 定义
var (
	ErrNoDefaultRoute                 = errors.New("no default route") // 没有找到 default 路由信息
	ErrNetworkOverlapsWithNameservers = errors.New("requested network overlaps with nameserver")
)
