package errcode

type ErrorCoder interface {
	ErrorCode() ErrorCode
}

type ErrorCode int

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
