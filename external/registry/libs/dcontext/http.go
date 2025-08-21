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
