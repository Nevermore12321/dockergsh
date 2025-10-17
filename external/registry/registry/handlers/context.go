package handlers

import (
	"context"
	"github.com/Nevermore12321/dockergsh/external/registry/libs/dcontext"
	"github.com/Nevermore12321/dockergsh/external/registry/registry/api/errcode"
	v1 "github.com/Nevermore12321/dockergsh/external/registry/registry/api/v1"
	"github.com/Nevermore12321/dockergsh/external/registry/registry/auth"
)

const (
	// userKey 用来从 context 中获取用户信息对象
	userKey = "auth.user"

	// userNameKey 用来从 context 中获取用户名
	userNameKey = "auth.user.name"
)

type Context struct {
	*App
	context.Context
	urlBuilder *v1.URLBuilder
	Errors     errcode.Errors
}

func getName(ctx context.Context) string {
	return dcontext.GetStringValue(ctx, "vars.name")
}

type UserInfoContext struct {
	context.Context
	user auth.UserInfo
}

func (uic UserInfoContext) Value(key interface{}) interface{} {
	switch key {
	case userKey:
		return uic.user
	case userNameKey:
		return uic.user.Name
	}

	// 如果都是不是，从 context 中返回 key 对应的 value
	return uic.Context.Value(key)
}

// WithUser 返回一个带有授权用户信息的新的 context
func WithUser(ctx context.Context, user auth.UserInfo) context.Context {
	return UserInfoContext{
		Context: ctx,
		user:    user,
	}
}

type ResourceContext struct {
	context.Context
	resources []auth.Resource
}

type resourceKey struct{}

func (rc ResourceContext) Value(key interface{}) interface{} {
	if key == (resourceKey{}) {
		return rc.resources
	}

	return rc.Context.Value(key)
}

// WithResources 返回一个带有授权资源信息的新的 context
func WithResources(ctx context.Context, resources []auth.Resource) context.Context {
	return ResourceContext{
		Context:   ctx,
		resources: resources,
	}
}
