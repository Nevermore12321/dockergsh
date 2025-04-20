package api_server

import (
	"expvar"
	"fmt"
	"github.com/Nevermore12321/dockergsh/internal/engine"
	"github.com/Nevermore12321/dockergsh/internal/utils"
	"github.com/Nevermore12321/dockergsh/pkg/version"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"net/http"
	"net/http/pprof"
	"os"
)

type HttpApiFunc func(eng *engine.Engine, version version.Version, w http.ResponseWriter, r *http.Request, vars map[string]string) error

// 为 server 创建路由
func createRouter(eng *engine.Engine, logging, enableCors bool, dockerVersion string) (*mux.Router, error) {
	router := mux.NewRouter()
	if os.Getenv(utils.DockergshDebug) != "" {
		AttachProfiler(router)
	}

	// 路由映射
	m := map[string]map[string]HttpApiFunc{
		"GET": {
			"/_ping": ping,
		},
	}

	for method, routes := range m {
		for route, fct := range routes {
			log.Debugf("Registering %s, %s", method, route)
			localRoute := route
			localFct := fct
			localMethod := method

		}
	}
	return router, nil
}

// Replicated from expvar.go as not public.
func expvarHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	fmt.Fprintf(w, "{\n")
	first := true
	expvar.Do(func(kv expvar.KeyValue) {
		if !first {
			fmt.Fprintf(w, ",\n")
		}
		first = false
		fmt.Fprintf(w, "%q: %s", kv.Key, kv.Value)
	})
	fmt.Fprintf(w, "\n}\n")
}

func AttachProfiler(router *mux.Router) {
	router.HandleFunc("/debug/vars", expvarHandler)                                         // 公共变量
	router.HandleFunc("/debug/pprof/", pprof.Index)                                         // 调试的首页
	router.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)                                // 当前程序的命令行的完整调用路径。
	router.HandleFunc("/debug/pprof/profile", pprof.Profile)                                // 默认进行 30s 的 CPU Profiling，得到一个分析用的 profile 文件
	router.HandleFunc("/debug/pprof/symbol", pprof.Symbol)                                  // 符号表
	router.HandleFunc("/debug/pprof/heap", pprof.Handler("heap").ServeHTTP)                 // 查看活动对象的内存分配情况
	router.HandleFunc("/debug/pprof/goroutine", pprof.Handler("goroutine").ServeHTTP)       // 查看当前所有运行的 goroutines 堆栈跟踪
	router.HandleFunc("/debug/pprof/threadcreate", pprof.Handler("threadcreate").ServeHTTP) // 查看创建新 OS 线程的堆栈跟踪
}
