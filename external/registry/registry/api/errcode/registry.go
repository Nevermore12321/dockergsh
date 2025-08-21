package errcode

import (
	"fmt"
	"sync"
)

var (
	registerLock sync.Mutex
	nextCode     = 1000 // 错误码从 1000 开始递增
)

const errGroup = "registry.api.v2"

func Register(group string, descriptor ErrorDescriptor) ErrorCode {
	return register(group, descriptor)
}

func register(group string, descriptor ErrorDescriptor) ErrorCode {
	registerLock.Lock()
	defer registerLock.Unlock()

	// 根据当前错误码序号 生成 新的错误码
	descriptor.Code = ErrorCode(nextCode)

	// 如果已经卸载，直接报错
	if _, ok := idToDescriptors[descriptor.Value]; ok {
		panic(fmt.Sprintf("ErrorValue %q is already registered", descriptor.Value))
	}

	if _, ok := errorCodeToDescriptors[descriptor.Code]; ok {
		panic(fmt.Sprintf("ErrorCode %v is already registered", descriptor.Code))
	}

	// 注册
	groupToDescriptors[group] = append(groupToDescriptors[group], descriptor)
	errorCodeToDescriptors[descriptor.Code] = descriptor
	idToDescriptors[descriptor.Value] = descriptor

	nextCode++
	return descriptor.Code
}
