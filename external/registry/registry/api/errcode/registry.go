package errcode

import (
	"net/http"
	"sync"
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

var (
	registerLock sync.Mutex
)

func Register() ErrorCode {

}
