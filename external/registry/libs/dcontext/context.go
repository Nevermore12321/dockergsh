package dcontext

import (
	"context"
	"sync"

	"github.com/google/uuid"
)

// 提供具有 instance id 的 context
type instanceContext struct {
	context.Context
	id   string    // instance id
	once sync.Once // 只在初始化时执行一次
}

// Value 第一次取 instance.id value 时，初始化写入 id 属性
func (ic *instanceContext) Value(key interface{}) interface{} {
	if key == "instance.id" { // 如果取 instance.id ，随机生成 id，写入
		ic.once.Do(func() {
			ic.id = uuid.NewString()
		})
		return ic.id
	}

	return ic.Context.Value(key)
}

var background = &instanceContext{
	Context: context.Background(),
}

func Background() context.Context {
	return background
}
