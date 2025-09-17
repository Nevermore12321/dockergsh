package handlers

import (
	"context"
	"github.com/Nevermore12321/dockergsh/external/registry/libs/dcontext"
	"github.com/Nevermore12321/dockergsh/external/registry/registry/api/errcode"
	v1 "github.com/Nevermore12321/dockergsh/external/registry/registry/api/v1"
	"github.com/Nevermore12321/dockergsh/external/registry/registry/auth"
	"net/http"
)

func (app *App) logError(ctx context.Context, errors errcode.Errors) {
	for _, err := range errors {
		var c context.Context

		switch e := err.(type) {
		case errcode.Error:
			c = context.WithValue(ctx, errCodeKey{}, e.Code)
			c = context.WithValue(c, errMessageKey{}, e.Message)
			c = context.WithValue(c, errDetailKey{}, e.Detail)
		case errcode.ErrorCode:
			c = context.WithValue(ctx, errCodeKey{}, e)
			c = context.WithValue(c, errMessageKey{}, e.Message())
		default:
			// just normal go 'error'
			c = context.WithValue(ctx, errCodeKey{}, errcode.ErrorCodeUnknown)
			c = context.WithValue(c, errMessageKey{}, e.Error())
		}

		c = dcontext.WithLogger(c, dcontext.GetLogger(c, errCodeKey{}, errMessageKey{}, errDetailKey{}))

		dcontext.GetResponseLogger(c).Errorf("response completed with error")
	}
}

func (app *App) context(w http.ResponseWriter, r *http.Request) *Context {
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

	if app.httpHost.Scheme != "" && app.httpHost.Host != "" {
		reqCtx.urlBuilder = v1.NewURLBuilder(&app.httpHost, false)
	} else {
		reqCtx.urlBuilder = v1.NewURLBuilderFromRequest(r, app.Config.HTTP.RelativeURLs)
	}

	return reqCtx
}

// 鉴权中间件
func (app *App) authorized(w http.ResponseWriter, r *http.Request, ctx Context) error {
	dcontext.GetLogger(ctx).Debug("authorizing request")
	// 获取 repository 名称
	repo := getName(ctx)

	// 如果没有 auth 插件，则直接跳过
	if app.accessController == nil {
		return nil
	}

	// 需要的授权列表
	var accessRecords []auth.Access
	if repo != "" {
		accessRecords = appendAccessRecords(accessRecords, r.Method, repo)

		// 将blob从一个存储库挂载到另一个存储库需要对源存储库进行pull （GET）访问。
		if fromRepo := r.FormValue("from"); fromRepo != "" {
			accessRecords = appendAccessRecords(accessRecords, http.MethodGet, fromRepo)
		}
	} else {
		// todo
	}

}

// 通过 request method 添加 Access 列表
func appendAccessRecords(records []auth.Access, method string, repo string) []auth.Access {
	// 当前是一个 repository 资源
	resource := auth.Resource{
		Type: "repository",
		Name: repo,
	}

	switch method {
	case http.MethodGet, http.MethodHead: // docker pull 权限
		records = append(records, auth.Access{
			Resource: resource,
			Action:   "pull",
		})
	case http.MethodPost, http.MethodPatch, http.MethodPut: // docker push 权限
		records = append(records,
			auth.Access{
				Resource: resource,
				Action:   "pull",
			},
			auth.Access{
				Resource: resource,
				Action:   "push",
			})
	case http.MethodDelete: // docker rmi 权限
		records = append(records,
			auth.Access{
				Resource: resource,
				Action:   "delete",
			})
	}
	return records

}
