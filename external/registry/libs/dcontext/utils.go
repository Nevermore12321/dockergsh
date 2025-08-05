package dcontext

import "context"

func GetStringValue(ctx context.Context, key interface{}) string {
	var value string
	if val, ok := ctx.Value(key).(string); ok {
		value = val
	}
	return value
}
