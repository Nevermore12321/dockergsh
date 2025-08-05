package dcontext

import "context"

type versionKey struct{}

func (versionKey) String() string {
	return "version"
}

// WithVersion ctx 中添加 version 信息
func WithVersion(ctx context.Context, version string) context.Context {
	ctx = context.WithValue(ctx, versionKey{}, version)

	// 使用默认的 logger，并且添加 version 的信息到 logger fields 中
	return WithLogger(ctx, GetLogger(ctx, versionKey{}))
}

// GetVersion 从 ctx 获取 version 信息
func GetVersion(ctx context.Context) string {
	return GetStringValue(ctx, versionKey{})
}
