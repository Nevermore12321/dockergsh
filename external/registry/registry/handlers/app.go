package handlers

import (
	"context"
	"github.com/Nevermore12321/dockergsh/external/registry/libs/configuration"
	"github.com/gorilla/mux"
)

//	App是一个全局注册应用对象。可以放置共享资源。所有请求都可以访问这个对象。
//
// 任何可写的字段应该被保护
type App struct {
	context.Context
	Config *configuration.Configuration // 配置文件
	Router *mux.Router                  // 主应用路由

	// todo
}

func NewApp(ctx context.Context, config *configuration.Configuration) *App {
	app := &App{
		Context: ctx,
		Config:  config,
		Router:  v2.Router,
	}
}
