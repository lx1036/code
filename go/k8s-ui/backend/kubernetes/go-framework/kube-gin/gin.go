package kube_gin

import (
	"net/http"
)

type HandlerFunc func(*Context)

type Engine struct {
	router *router
	*RouterGroup
}

func (engine *Engine) addRoute(method string, pattern string, handler HandlerFunc) {
	engine.router.addRoute(method, pattern, handler)
}

func (engine *Engine) Get(pattern string, handler HandlerFunc) {
	engine.router.addRoute("GET", pattern, handler)
}
func (engine *Engine) Post(pattern string, handler HandlerFunc) {
	engine.router.addRoute("POST", pattern, handler)
}

func (engine *Engine) Group(prefix string) *RouterGroup {
	engine.RouterGroup.prefix = prefix
	return engine.RouterGroup
}

func (engine *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	ctx := newContext(w, req)
	engine.router.handle(ctx)
}

func (engine *Engine) Run(addr string) error {
	return http.ListenAndServe(addr, engine)
}

func New() *Engine {
	router := newRouter()
	return &Engine{
		router:      router,
		RouterGroup: &RouterGroup{router: router},
	}
}
