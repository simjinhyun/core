package x

import "net/http"

// Context : 요청/응답을 담는 컨텍스트
type Context struct {
	Req *http.Request
	Res http.ResponseWriter
}

func NewContext(w http.ResponseWriter, r *http.Request) *Context {
	return &Context{Req: r, Res: w}
}
