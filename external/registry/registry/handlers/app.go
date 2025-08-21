package handlers

import (
	"context"
	"github.com/Nevermore12321/dockergsh/external/registry/libs/configuration"
	"github.com/Nevermore12321/dockergsh/external/registry/libs/dcontext"
	v1 "github.com/Nevermore12321/dockergsh/external/registry/registry/api/v1"
	"github.com/gorilla/mux"
	"net/http"
)

//	App是一个全局注册应用对象。可以放置共享资源。所有请求都可以访问这个对象。
//
// 任何可写的字段应该被保护
type App struct {
	context.Context
	Config *configuration.Configuration // 配置文件
	Router *mux.Router                  // 主应用路由

	isCahce bool // 是否开启缓存
	// todo
}

func NewApp(ctx context.Context, config *configuration.Configuration) *App {
	app := &App{
		Context: ctx,
		Config:  config,
		Router:  v1.RouterWithPrefix(config.HTTP.Prefix),
		isCahce: config.Proxy.RemoteURL != "",
	}

	// 注册路由 handler
	app.register()
}

type dispatchFunc func(ctx *Context, r *http.Request) http.Handler

func (app *App) register(routeName string, dispatchFunc dispatchFunc) {

}

// 生成真正的 http handler
func (app *App) dispatcher(dispatchFunc2 dispatchFunc) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		// 设置配置的 公共响应 header
		for headerName, headerValue := range app.Config.HTTP.Headers {
			for _, value := range headerValue {
				writer.Header().Add(headerName, value)
			}
		}

		ctx := app.con
	})
}

func (app *App) context(w http.ResponseWriter, r http.Request) *Context {
	ctx := r.Context()
	ctx = dcontext.WithVars(ctx, r)
	ctx = dcontext.WithLogger(ctx, dcontext.GetLogger(ctx,
		"vars.name",
		"vars.reference",
		"vars.digest",
		"vars.uuid",
	))

	reqCtx := &Context{
		App:     app,
		Context: ctx,
	}

	return reqCtx
}
