package x

import (
	"net/http"
	"os"
	"path/filepath"
)

type HandlerFunc func(*Context)

type Route struct {
	Path     string
	Method   string
	Type     string
	Handlers []HandlerFunc
}

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

// JSON 응답용 라우트 등록
func (r *Router) HandleJSON(method, path string, handlers ...HandlerFunc) {
	key := method + " " + path

	if _, ok := r.routes[key]; ok {
		panic("router: route already registered: " + key)
	}

	r.routes[key] = &Route{
		Path:     path,
		Method:   method,
		Type:     "JSON",
		Handlers: handlers,
	}
}

// HTML 응답용 라우트 등록
func (r *Router) HandleHTML(method, path string, handlers ...HandlerFunc) {
	key := method + " " + path

	if _, ok := r.routes[key]; ok {
		panic("router: route already registered: " + key)
	}

	r.routes[key] = &Route{
		Path:     path,
		Method:   method,
		Type:     "HTML",
		Handlers: handlers,
	}
}

func (r *Router) ServeHTTP(c *Context) {
	key := c.Req.Method + " " + c.Req.URL.Path

	if route, ok := r.routes[key]; ok {
		c.RouteType = route.Type
		// 등록된 핸들러들을 순서대로 실행
		for _, h := range route.Handlers {
			h(c)
		}
		return
	}

	// 등록된 라우트가 없으면 정적 파일 제공
	path := filepath.Join(r.WebRoot, c.Req.URL.Path)
	http.ServeFile(c.Res, c.Req, path)
}
