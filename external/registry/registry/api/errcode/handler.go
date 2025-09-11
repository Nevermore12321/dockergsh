package errcode

import (
	"encoding/json"
	"net/http"
)

// ServeJson 如果 errors 存在，直接在响应中返回 错误信息
func ServeJson(w http.ResponseWriter, err error) error {
	w.Header().Set("Content-Type", "application/json")

	var sc int
	switch errs := err.(type) {
	case Errors:
		if errs.Len() < 1 {
			break
		}

		if err, ok := errs[0].(ErrorCoder); ok {
			sc = err.ErrorCode().Descriptor().HTTPStatusCode
		}
	case ErrorCoder:
		sc = errs.ErrorCode().Descriptor().HTTPStatusCode
		err = Errors{err}
	default:
		err = Errors{err}
	}

	if sc == 0 {
		sc = http.StatusInternalServerError
	}

	w.WriteHeader(sc)

	return json.NewEncoder(w).Encode(err)
}
