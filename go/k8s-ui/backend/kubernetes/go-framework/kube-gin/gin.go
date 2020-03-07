package kube_gin

import (
	"net/http"
)

type HandlerFunc func(*Context)

type Engine struct {
	routers *router
}

func (engine *Engine) addRoute(method string, pattern string, handler HandlerFunc) {
	engine.routers.addRoute(method, pattern, handler)
}

func (engine *Engine) Get(pattern string, handler HandlerFunc) {
	engine.routers.addRoute("GET", pattern, handler)
}
func (engine *Engine) Post(pattern string, handler HandlerFunc) {
	engine.routers.addRoute("POST", pattern, handler)
}

func (engine *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	ctx := newContext(w, req)
	engine.routers.handle(ctx)
}

func (engine *Engine) Run(addr string) error {
	return http.ListenAndServe(addr, engine)
}

func New() *Engine {
	return &Engine{
		routers: newRouter(),
	}
}
