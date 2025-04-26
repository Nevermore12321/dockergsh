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
	"strings"
)

type HttpApiFunc func(eng *engine.Engine, version version.Version, w http.ResponseWriter, r *http.Request, vars map[string]string) error

// 通过 HttpApiFunc 转成 http handler
func makeHttpHandler(eng *engine.Engine, logging bool, localMethod, localRoute string, apiFunc HttpApiFunc, enableCors bool, dockerVersion version.Version) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 记录日志
		log.Debugf("Calling %s %s", localMethod, localRoute)
		if logging {
			log.Infof("%s %s", localMethod, localRoute)
		}

		// dockergsh-client 调用时，需要与服务端的版本一致
		if strings.Contains(r.Header.Get("User-Agent"), "Dockergsh-Client/") {
			userAgent := strings.Split(r.Header.Get("User-Agent"), "/")
			if len(userAgent) == 2 && !dockerVersion.Equal(version.Version(userAgent[1])) {
				log.Debugf("Warning: client and server don't have the same version (client: %s, server: %s)", userAgent[1], dockerVersion)
			}
		}

		// 从路由中获取 version 的路由变量
		pathVersion := version.Version(mux.Vars(r)["version"])
		if pathVersion == "" {
			pathVersion = APIVERSION
		}

		if enableCors {
			writeCorsHeaders(w)
		}
		if pathVersion.GreaterThan(APIVERSION) {
			http.Error(w, fmt.Errorf("client and server don't have same version (client : %s, server: %s)", pathVersion, APIVERSION).Error(), http.StatusNotFound)
			return
		}

		// 执行 handler 函数
		if err := apiFunc(eng, pathVersion, w, r, mux.Vars(r)); err != nil {
			log.Errorf("Handler for %s %s returned error: %s", localMethod, localRoute, err)
			httpError(w, err)
		}
	}
}

// response 写入跨域相关的 header
func writeCorsHeaders(w http.ResponseWriter) {
	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept")
	w.Header().Add("Access-Control-Allow-Methods", "GET, POST, DELETE, PUT, OPTIONS")
}

// 根据返回的 error 信息，决定返回什么 http status code
func httpError(w http.ResponseWriter, err error) {
	statusCode := http.StatusInternalServerError
	if strings.Contains(err.Error(), "No such") {
		statusCode = http.StatusNotFound
	} else if strings.Contains(err.Error(), "Bad parameter") {
		statusCode = http.StatusBadRequest
	} else if strings.Contains(err.Error(), "Conflict") {
		statusCode = http.StatusConflict
	} else if strings.Contains(err.Error(), "Impossible") {
		statusCode = http.StatusNotAcceptable
	} else if strings.Contains(err.Error(), "Wrong login/password") {
		statusCode = http.StatusUnauthorized
	} else if strings.Contains(err.Error(), "hasn't been activated") {
		statusCode = http.StatusForbidden
	}

	if err.(error) != nil {
		log.Errorf("HTTP Error: statusCode=%d %s", statusCode, err.Error())
		http.Error(w, err.Error(), statusCode)
	}
}

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
			"/info":  getInfo,
		},
	}

	for method, routes := range m {
		for route, fct := range routes {
			log.Debugf("Registering %s, %s", method, route)
			localRoute := route
			localFct := fct
			localMethod := method

			// 注册 handler
			handler := makeHttpHandler(eng, logging, localMethod, localRoute, localFct, enableCors, version.Version(dockerVersion))

			// 添加路由
			if localRoute == "" {
				// 如果是根不添加 /vx.x.x version path
				router.Methods(localMethod).HandlerFunc(handler)
			} else { // 添加两个路由
				// 添加 version path
				router.Path("/v{version:[0-9.]+}" + localRoute).Methods(localMethod).HandlerFunc(handler)
				// 不添加 version path
				router.Path(localRoute).Methods(localMethod).HandlerFunc(handler)
			}
		}
	}
	return router, nil
}

// expvar 包为公共变量提供了一个标准化的接口，如服务接口中的访问计数器。
// 输出所有公共变量的值
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
