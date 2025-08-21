package errcode

import "net/http"

var (
	errorCodeToDescriptors = map[ErrorCode]ErrorDescriptor{} // 错误码 - 错误描述符映射
	idToDescriptors        = map[string]ErrorDescriptor{}    // 错误id - 错误描述符映射
	groupToDescriptors     = map[string][]ErrorDescriptor{}  // 错误组 - 包含那些错误描述符
)

var (
	ErrorCodeTooManyRequests = register("errcode", ErrorDescriptor{
		Value:   "TOOMANYREQUESTS",
		Message: "too many requests",
		Description: `Returned when a client attempts to contact a
		service too many times`,
		HTTPStatusCode: http.StatusTooManyRequests,
	})
)
