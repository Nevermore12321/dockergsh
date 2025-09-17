package handlers

import (
	"context"
	"fmt"
	"github.com/Nevermore12321/dockergsh/external/registry/libs/configuration"
	"github.com/Nevermore12321/dockergsh/external/registry/libs/dcontext"
	"github.com/Nevermore12321/dockergsh/external/registry/registry/api/errcode"
	v1 "github.com/Nevermore12321/dockergsh/external/registry/registry/api/v1"
	"github.com/Nevermore12321/dockergsh/external/registry/registry/auth"
	"github.com/gorilla/mux"
	"net/http"
	"net/url"
)

//	App是一个全局注册应用对象。可以放置共享资源。所有请求都可以访问这个对象。
//
// 任何可写的字段应该被保护
type App struct {
	context.Context
	Config *configuration.Configuration // 配置文件
	Router *mux.Router                  // 主应用路由

	httpHost url.URL // 表示 http.Host

	isCahce bool // 是否开启缓存

	accessController auth.AccessController // 鉴权
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
	app.register(v1.RouteNameBase, func(ctx *Context, r *http.Request) http.Handler {
		return http.HandlerFunc(apiBase)
	})
}

// 返回简单的 /v1/ ，返回空
func apiBase(w http.ResponseWriter, r *http.Request) {
	const emptyJSON = "{}"

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Length", fmt.Sprint(len(emptyJSON)))

	fmt.Fprint(w, emptyJSON)
}

type dispatchFunc func(ctx *Context, r *http.Request) http.Handler

// 注册 handler 到 router
func (app *App) register(routeName string, dispatchFunc dispatchFunc) {
	// 通过 dispatch 获取到 handler
	handler := app.dispatcher(dispatchFunc)

	// 如果开启 Prometheus
	if app.Config.HTTP.Debug.Prometheus.Enabled {
		// todo
	}

	app.Router.GetRoute(routeName).Handler(handler)
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

		ctx := app.context(writer, request)

		// 自动处理 错误，返回 错误 response
		defer func() {
			if ctx.Errors.Len() > 0 { // 有错误时，返回错误，并打印日志
				_ = errcode.ServeJson(writer, ctx.Errors)
				app.logError(ctx, ctx.Errors)
			} else if status, ok := ctx.Value("http.response.status").(int); ok && status >= 200 && status < 399 { // 成功，并打印日志
				dcontext.GetResponseLogger(ctx).Infof("response completed")
			}
		}()

		// 鉴权中间件
		if err := app.authorized(); err != nil {
			dcontext.GetLogger(ctx).Warnf("error authorizing context: %v", err)
			return
		}
	})
}

type errCodeKey struct{}

func (errCodeKey) String() string { return "err.code" }

type errMessageKey struct{}

func (errMessageKey) String() string { return "err.message" }

type errDetailKey struct{}

func (errDetailKey) String() string { return "err.detail" }
