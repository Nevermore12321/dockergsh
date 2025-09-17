package dcontext

import (
	"context"
	"github.com/gorilla/mux"
	"net/http"
)

type muxVarsContext struct {
	context.Context
	vars map[string]string
}

// WithVars mux.Vars(r)获取请求r的路由变量，可以通过此 context 获取相关变量
func WithVars(ctx context.Context, r *http.Request) context.Context {
	return &muxVarsContext{
		Context: ctx,
		vars:    mux.Vars(r), // 从 request 中获取 vars 路由变量放入 context
	}
}

// GetRequestLogger 通过 request 添加 日志记录
func GetRequestLogger(ctx context.Context) Logger {
	return GetLogger(ctx,
		"http.request.id",
		"http.request.method",
		"http.request.host",
		"http.request.uri",
		"http.request.referer",
		"http.request.useragent",
		"http.request.remoteaddr",
		"http.request.contenttype")
}

// GetResponseLogger  通过 response 添加 日志记录
func GetResponseLogger(ctx context.Context) Logger {
	l := getLogrusLogger(ctx,
		"http.response.written",
		"http.response.status",
		"http.response.contenttype")

	// 计算响应时间
	duration := Since(ctx, "http.request.startedat")

	if duration > 0 {
		l = l.WithField("http.response.duration", duration.String())
	}
	return l
}
