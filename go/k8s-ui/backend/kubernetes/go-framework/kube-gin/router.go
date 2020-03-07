package kube_gin

import (
	"fmt"
	"net/http"
)

type router struct {
	handlers map[string]map[string]HandlerFunc
}

func newRouter() *router {
	return &router{handlers: map[string]map[string]HandlerFunc{}}
}

func (r *router) addRoute(method string, pattern string, handler HandlerFunc) {
	if r.handlers[method] != nil {

	} else {
		r.handlers[method] = map[string]HandlerFunc{}

	}

	r.handlers[method][pattern] = handler
}

func (r *router) handle(ctx *Context) {
	var patternHandler map[string]HandlerFunc
	patternHandler, ok := r.handlers[ctx.Method]
	if !ok {
		ctx.Status(http.StatusBadRequest)
		_, _ = fmt.Fprintf(ctx.Writer, "Bad request: %s", ctx.Req.URL)
		return
	}

	handler, ok := patternHandler[ctx.Path]
	if !ok {
		ctx.Status(http.StatusBadRequest)
		_, _ = fmt.Fprintf(ctx.Writer, "Bad request: %s", ctx.Req.URL)
		return
	}

	handler(ctx)
}
