package x

import (
	"net/http"
	"path/filepath"
)

type HandlerFunc func(*Context)

type Route struct {
	Path    string
	Type    string // "json" or "html"
	Handler HandlerFunc
}

type Router struct {
	routes map[string]*Route
}

var router *Router

func init() {
	router = NewRouter()
}

func GetRouter() *Router {
	return router
}

func NewRouter() *Router {
	return &Router{routes: make(map[string]*Route)}
}

func (r *Router) HandleJSON(path string, h HandlerFunc) {
	r.routes[path] = &Route{Path: path, Type: "json", Handler: h}
}

func (r *Router) HandleHTML(path string, h HandlerFunc) {
	r.routes[path] = &Route{Path: path, Type: "html", Handler: h}
}

func (r *Router) ServeHTTP(c *Context) {
	if route, ok := r.routes[c.Req.URL.Path]; ok {
		route.Handler(c)
		return
	}

	// App 설정에서 WebRoot 가져오기
	root := c.App.Conf["WebRoot"].(string)
	path := filepath.Join(root, c.Req.URL.Path)

	http.ServeFile(c.Res, c.Req, path)
}
