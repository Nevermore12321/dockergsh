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

// Challenge 是一种特殊的错误类型，用于 HTTP 401 未经授权的响应，并且能够根据错误编写带有WWW-Authenticate header 的响应。
type Challenge interface {
	error

	// SetHeaders 当遇到 401 错误时，可以使用此方法对响应添加响应的 header
	SetHeaders(r *http.Request, w http.ResponseWriter)
}
