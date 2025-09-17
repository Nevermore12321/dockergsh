package auth

import "net/http"

// Resource 根据类型和名称定义一个资源
type Resource struct {
	Type  string
	Class string
	Name  string
}

type Access struct {
	Resource
	Action string
}

type UserInfo struct {
	Name string // 用户名
}

// Grant 鉴权信息
type Grant struct {
	User      UserInfo   // 用户信息
	Resources []Resource // 用户授权的资源列表
}

type AccessController interface {
	Authorized(r *http.Request, access ...Access) (*Grant, error)
}
