package api_server

import (
	"github.com/Nevermore12321/dockergsh/internal/engine"
	"net/http"
)

func ping(eng *engine.Engine, version version.Version, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
}
