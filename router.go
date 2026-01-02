package x

import (
	"net/http"
	"os"
	"path/filepath"
)

type HandlerFunc func(*Context)

type Route struct {
	Path    string
	Type    string // "json" or "html"
	Handler HandlerFunc
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

func (r *Router) HandleJSON(path string, h HandlerFunc) {
	r.routes[path] = &Route{Path: path, Type: "json", Handler: h}
}

func (r *Router) HandleHTML(path string, h HandlerFunc) {
	r.routes[path] = &Route{Path: path, Type: "html", Handler: h}
}

func (r *Router) ServeHTTP(c *Context) {
	c.App.Logger.Debug("[REQ]", "Path", c.Req.URL.Path)
	if route, ok := r.routes[c.Req.URL.Path]; ok {
		route.Handler(c)
		return
	}
	path := filepath.Join(r.WebRoot, c.Req.URL.Path)
	http.ServeFile(c.Res, c.Req, path)
}
