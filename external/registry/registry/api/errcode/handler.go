package errcode

import "net/http"

// ServeJson 如果 errors 存在，直接在响应中返回 错误信息
func ServeJson(w http.ResponseWriter, err error) error {
	w.Header().Set("Content-Type", "application/json")

	var sc int
	switch errs := err.(type) {
	case Errors:
		if errs.Len() < 1 {
			break
		}
		// todo
	}
}
