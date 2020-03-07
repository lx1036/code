package kube_gin

import (
	"fmt"
	"net/http"
)

type Engine struct {
	routers map[string]map[string]http.HandlerFunc
}

func (engine *Engine) addRoute(method string, pattern string, handler http.HandlerFunc) {
	if engine.routers[method] != nil {

	} else {
		engine.routers[method] = map[string]http.HandlerFunc{}
	}

	engine.routers[method][pattern] = handler
}

func (engine *Engine) Get(pattern string, handler http.HandlerFunc) {
	engine.addRoute("GET", pattern, handler)
}
func (engine *Engine) Post(pattern string, handler http.HandlerFunc) {
	engine.addRoute("POST", pattern, handler)
}

func (engine *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var patternHandler map[string]http.HandlerFunc
	patternHandler, ok := engine.routers[req.Method]
	if !ok {
		_, _ = fmt.Fprintf(w, "Bad request: %s", req.URL)
	}

	handler, ok := patternHandler[req.URL.Path]
	if !ok {
		_, _ = fmt.Fprintf(w, "Bad request: %s", req.URL)
	}

	handler(w, req)
}

func (engine *Engine) Run(addr string) error {
	return http.ListenAndServe(addr, engine)
}

func New() *Engine {
	return &Engine{
		routers: make(map[string]map[string]http.HandlerFunc),
	}
}
