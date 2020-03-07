package kube_gin

import (
	"net/http"
	"strings"
)

type router struct {
	roots    map[string]*node
	handlers map[string]map[string]HandlerFunc
}

func newRouter() *router {
	return &router{
		roots:    map[string]*node{},
		handlers: map[string]map[string]HandlerFunc{},
	}
}

func parsePattern(pattern string) []string {
	parts := strings.Split(pattern, "/")
	var results []string
	for _, part := range parts {
		if part != "" {
			results = append(results, part)
			if part[0] == '*' { // /person/:name/*
				break
			}
		}
	}

	return results
}

func (r *router) addRoute(method string, pattern string, handler HandlerFunc) {
	parts := parsePattern(pattern)
	_, ok := r.roots[method]
	if !ok {
		r.roots[method] = &node{}
	}

	r.roots[method].insert(pattern, parts, 0)

	if _, ok := r.handlers[method]; !ok {
		r.handlers[method] = map[string]HandlerFunc{}
	}

	r.handlers[method][pattern] = handler
}

func (r *router) getRoute(method string, pattern string) (*node, map[string]string) {
	searchParts := parsePattern(pattern) // /people/1/accounts
	root, ok := r.roots[method]
	if !ok {
		return nil, nil
	}

	node := root.search(searchParts, 0)
	if node != nil {
		params := map[string]string{}       // 提取出动态路由参数
		parts := parsePattern(node.pattern) // /people/:id/accounts
		for index, part := range parts {
			if part[0] == ':' {
				params[part[1:]] = searchParts[index] // id=1
			}
			if part[0] == '*' && len(part) > 1 { // /people/*id/accounts -> /people/1/accounts
				params[part[1:]] = strings.Join(searchParts[index:], "/") // id="1/accounts"
				break
			}
		}

		return node, params
	}

	return nil, nil
}

func (r *router) handle(ctx *Context) {
	route, params := r.getRoute(ctx.Method, ctx.Path)
	if route != nil {
		ctx.Params = params
		handler := r.handlers[ctx.Method][route.pattern]
		if handler != nil {
			handler(ctx)
		} else {
			ctx.JSON(http.StatusInternalServerError, H{
				"errno":  -1,
				"errmsg": "not found matched route handler",
			})
		}
	} else {
		ctx.JSON(http.StatusNotFound, H{
			"errno":  -1,
			"errmsg": "not found",
		})
	}
}
