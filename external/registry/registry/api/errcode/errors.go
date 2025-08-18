package errcode

type ErrorCoder interface {
	ErrorCode() ErrorCode
}

type ErrorCode int
