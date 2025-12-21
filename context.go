package x

import "net/http"

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

func (c *Context) Reply() {

}
