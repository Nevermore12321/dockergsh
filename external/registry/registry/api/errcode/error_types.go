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

	// ErrorCodeUnknown 未知错误
	ErrorCodeUnknown = register("errcode", ErrorDescriptor{
		Value:   "UNKNOWN",
		Message: "unknown error",
		Description: `Generic error returned when the error does not have an
			                                            API classification.`,
		HTTPStatusCode: http.StatusInternalServerError,
	})

	// ErrorCodeUnauthorized 未认证错误
	ErrorCodeUnauthorized = register("errcode", ErrorDescriptor{
		Value:   "UNAUTHORIZED",
		Message: "authentication required",
		Description: `The access controller was unable to authenticate
		the client. Often this will be accompanied by a
		Www-Authenticate HTTP response header indicating how to
		authenticate.`,
		HTTPStatusCode: http.StatusUnauthorized,
	})
)
