package v1

import (
	"github.com/gorilla/mux"
	"sync"
)

var (
	baseRouter           *mux.Router
	createBaseRouterOnce sync.Once
)

func Router() *mux.Router {
	createBaseRouterOnce.Do(func() {
		baseRouter = RouterWithPrefix("")
	})
	return baseRouter
}

func RouterWithPrefix(prefix string) *mux.Router {
	rootRouter := mux.NewRouter()
	router := rootRouter
	if prefix != "" {
		router = router.PathPrefix(prefix).Subrouter()
	}

	router.StrictSlash(true)

	for _, descriptor := range routeDescriptors {
		router.Path(descriptor.Path).Name(descriptor.Name)
	}

	return router
}
