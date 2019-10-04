package router

import "net/http"

type Params []Param

type Param struct {
	Key string
	Value string
}

func (param Param) ByName(key string) string  {
	if param.Key == key {
		return param.Value
	}

	return ""
}

func (params Params) ByName(key string) string  {
	for _, param := range params {
		if param.Key == key {
			return param.Value
		}
	}

	return ""
}

type Handle func(http.ResponseWriter, *http.Request, Params)

type Route struct {
	path string
	handle Handle
}





type RouteCollection struct {
	routes map[string]Handle
}

func (routeCollection *RouteCollection)addRoute(path string, handle Handle)  {
	routeCollection.routes[path] = handle
}

func (routeCollection *RouteCollection) getRoutes(path string) (handle Handle, params Params)  {
	handle = routeCollection.routes[path]

	return handle, nil
}

type Router struct {
	routes map[string]*RouteCollection
	NotFound http.Handler
}

func (router *Router) Handle(method string, path string, handle Handle)  {
	if router.routes == nil {
		router.routes = make(map[string]*RouteCollection)
	}

	routeCollection := router.routes[method]
	if routeCollection == nil {
		routeCollection = new(RouteCollection)
		router.routes[method] = routeCollection
	}

	routeCollection.addRoute(path, handle)
}

func (router *Router) ServeHTTP(writer http.ResponseWriter, request *http.Request)  {
	path := request.URL.Path
	if routeCollection := router.routes[request.Method]; routeCollection != nil {
		if handle, params :=routeCollection.getRoutes(path); handle != nil {
			handle(writer, request, params)
			return
		} else {

		}
	}

	// 404
	if router.NotFound != nil {

	} else {

	}
}

func New() *Router  {
	return &Router{}
}
