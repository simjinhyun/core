package x

import (
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
)

type Router struct {
	WebRoot string
	routes  map[string]*Route
}

func NewRouter(WebRoot string) *Router {
	if WebRoot == "" {
		WebRoot = "./www"
	}
	if err := os.MkdirAll(WebRoot, 0755); err != nil {
		panic(err)
	}
	return &Router{
		WebRoot: WebRoot,
		routes:  make(map[string]*Route),
	}
}

type HandlerFunc func(*Context)

type Route struct {
	Path         string
	Method       string
	Reply        HandlerFunc
	Handlers     []HandlerFunc
	HandlerNames []string
}

func (r *Router) AddRoute(method, path string, reply HandlerFunc, handlers ...HandlerFunc) {
	names := make([]string, len(handlers))
	for i, h := range handlers {
		names[i] = runtime.FuncForPC(reflect.ValueOf(h).Pointer()).Name()
	}
	r.routes[method+" "+path] = &Route{
		Path:         path,
		Method:       method,
		Reply:        reply,
		Handlers:     handlers,
		HandlerNames: names,
	}
}

func (r *Router) ServeHTTP(c *Context) {
	key := c.Req.Method + " " + c.Req.URL.Path

	if route, ok := r.routes[key]; ok {
		// 등록된 핸들러들을 순서대로 실행
		for i, h := range route.Handlers {
			c.App.Logger.Debug(c.PrependXReqID("CALL " + route.HandlerNames[i]))
			h(c)
		}
		return
	}

	// 등록된 라우트가 없으면 정적 파일 제공
	path := filepath.Join(r.WebRoot, c.Req.URL.Path)
	http.ServeFile(c.Res, c.Req, path)
}
