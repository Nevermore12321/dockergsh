package dcontext

import (
	"context"
	"time"
)

func GetStringValue(ctx context.Context, key interface{}) string {
	var value string
	if val, ok := ctx.Value(key).(string); ok {
		value = val
	}
	return value
}

// Since 根据健计算请求的响应时间
func Since(ctx context.Context, key interface{}) time.Duration {
	if startAt, ok := ctx.Value(key).(time.Time); ok {
		return time.Since(startAt)
	}
	return 0
}
