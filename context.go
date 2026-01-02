package x

import (
	"fmt"
	"net/http"
)

// Context : 요청/응답을 담는 컨텍스트
type Context struct {
	App      *App
	Req      *http.Request
	Res      http.ResponseWriter
	AppError *AppError
	Response struct {
		Code    string `json:"Code"`
		Message string `json:"Message"`
		Data    any    `json:"Data"`
		Elapsed int64  `json:"Elapsed"`
	}
}

func NewContext(a *App, w http.ResponseWriter, r *http.Request) *Context {
	return &Context{App: a, Req: r, Res: w}
}

func (c *Context) Recover() {
	if rec := recover(); rec != nil {
		var appErr *AppError
		switch e := rec.(type) {
		case *AppError:
			appErr = e
		case error:
			appErr = NewAppError("RuntimeError", e, nil)
		default:
			appErr = NewAppError("RuntimeError", fmt.Errorf("%v", rec), nil)
		}
		c.AppError = appErr
	} else {
		c.AppError = NewAppError("OK", nil, nil)
	}
	c.Reply()
}

func (c *Context) Reply() {
	// 응답 직렬화/쓰기 로직은 여기서 처리
}
