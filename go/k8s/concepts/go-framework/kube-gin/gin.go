package kube_gin

import (
	"net/http"
	"sync"
)

type HandlerFunc func(*Context)
type HandlersChain []HandlerFunc

type Engine struct {
	*router
	//*RouterGroup
	groups []*RouterGroup

	RouterGroup

	HandleMethodNotAllowed bool
	ForwardedByClientIP    bool
	pool                   sync.Pool
	trees                  methodTrees
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

func (engine *Engine) Group(prefix string) RouterGroup {
	engine.RouterGroup.prefix = prefix
	return engine.RouterGroup
}

func (engine *Engine) allocateContext() *Context {
	return &Context{engine: engine}
}

func (engine *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	//ctx := newContext(w, req)
	//ctx.handlers = engine.RouterGroup.middlewares

	ctx := engine.pool.Get().(*Context)

	httpMethod := ctx.Request.Method
	path := ctx.Request.URL.Path

	tree := engine.trees
	for i, tl := 0, len(tree); i < tl; i++ {
		if tree[i].method != httpMethod {
			continue
		}

		//root := tree[i].root
		node, params := engine.router.getRoute(httpMethod, path)

	}

	engine.router.handle(ctx)

	engine.pool.Put(ctx)
}

func (engine *Engine) Run(addr string) error {
	return http.ListenAndServe(addr, engine)
}

func New() *Engine {
	/*engine := &Engine{
		router:      newRouter(),
	}
	engine.RouterGroup = &RouterGroup{engine:engine}
	engine.groups = []*RouterGroup{engine.RouterGroup}
	*/

	engine := &Engine{
		router: nil,
		RouterGroup: RouterGroup{
			Handlers: nil,
			basePath: "/",
			root:     true,
		},
		ForwardedByClientIP: true,
		groups:              nil,
		trees:               make(methodTrees, 0, 9),
	}

	engine.RouterGroup.engine = engine

	engine.pool.New = func() interface{} {
		return engine.allocateContext()
	}

	return engine
}

func Default() *Engine {
	engine := New()
	engine.Use(Logger(), Recovery())
	return engine
}
