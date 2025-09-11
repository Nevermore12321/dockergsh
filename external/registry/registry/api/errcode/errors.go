package errcode

import (
	"strings"
)

type ErrorCoder interface {
	ErrorCode() ErrorCode
}

type ErrorCode int

func (ec ErrorCode) ErrorCode() ErrorCode {
	return ec
}

func (ec ErrorCode) String() string {
	return ec.Descriptor().Value
}

func (ec ErrorCode) Error() string {
	return strings.ToLower(strings.Replace(ec.String(), "_", " ", -1))
}

func (ec ErrorCode) Descriptor() ErrorDescriptor {
	d, ok := errorCodeToDescriptors[ec]

	if !ok {
		return ErrorCodeUnknown.Descriptor()
	}

	return d
}

// ErrorDescriptor 提供 error 的详细信息
type ErrorDescriptor struct {
	// 该错误描述符的错误码
	Code ErrorCode

	// 错误码对应的错误类型值
	Value string

	// 错误信息
	Message string

	// 完整的错误描述
	Description string

	// http 状态码
	HTTPStatusCode int
}

type Errors []error

func (errs Errors) Len() int {
	return len(errs)
}

func (errs Errors) Error() string {
	switch len(errs) {
	case 0:
		return "<nil>"
	case 1:
		return errs[0].Error()
	default:
		msg := "errors: \n"
		for _, err := range errs {
			msg += err.Error() + "\n"
		}
		return msg
	}
}

//type Error struct {
//	Code    ErrorCode   `json:"code"`
//	Message string      `json:"message"`
//	Detail  interface{} `json:"detail,omitempty"`
//}
//
//func (e Error) Error() string {
//	return fmt.Sprintf("%s: %s", e.Code.Error(), e.Message)
//}
