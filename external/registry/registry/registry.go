package registry

import (
	"context"
	"github.com/Nevermore12321/dockergsh/external/registry/libs/configuration"
	"github.com/Nevermore12321/dockergsh/external/registry/registry/handlers"
	"net/http"
	"os"
)

type Registry struct {
	config *configuration.Configuration
	server *http.Server
	app    *handlers.App
	quit   chan os.Signal
}

// NewRegistry 根据配置文件和 context 生成 Registry 实例
func NewRegistry(ctx context.Context, config *configuration.Configuration) (*Registry, error) {
	var err error
	app := handlers.NewApp(ctx, config)

}
