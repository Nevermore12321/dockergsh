package v1

import (
	"github.com/gorilla/mux"
	"net/http"
	"net/url"
	"strings"
)

type URLBuilder struct {
	root     *url.URL // root url
	router   *mux.Router
	relative bool
}

func NewURLBuilder(root *url.URL, relative bool) *URLBuilder {
	return &URLBuilder{
		root:     root,
		router:   Router(),
		relative: relative,
	}
}

// NewURLBuilderFromRequest 从 request 中获取 root url 构建
func NewURLBuilderFromRequest(r *http.Request, relative bool) *URLBuilder {
	var (
		scheme = "http"
		host   = r.Host
	)

	if r.TLS != nil {
		scheme = "https"
	} else if len(r.URL.Scheme) > 0 {
		scheme = r.URL.Scheme
	}

	// 添加 x-forwarded-for 头
	if forwarded := r.Header.Get("Forwarded"); len(forwarded) > 0 {
		forwardedHeader, _, err := parseForwardedHeader(forwarded)
		if err == nil {
			if fproto := forwardedHeader["proto"]; len(fproto) > 0 {
				scheme = fproto
			}
			if fhost := forwardedHeader["host"]; len(fhost) > 0 {
				host = fhost
			}
		}
	} else {
		if forwardedProto := r.Header.Get("X-Forwarded-Proto"); len(forwardedProto) > 0 {
			scheme = forwardedProto
		}
		if forwardedHost := r.Header.Get("X-Forwarded-Host"); len(forwardedHost) > 0 {
			host, _, _ = strings.Cut(forwardedHost, ",")
			host = strings.TrimSpace(host)
		}
	}

	basePath := routeDescriptorsMap[RouteNameBase].Path

	reqPath := r.URL.Path

	index := strings.Index(reqPath, basePath)

	u := &url.URL{
		Scheme: scheme,
		Host:   host,
	}

	if index > 0 {
		u.Path = reqPath[0 : index+1]
	}

	return NewURLBuilder(u, relative)
}
