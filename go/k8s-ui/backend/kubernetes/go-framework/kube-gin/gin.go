package kube_gin

import (
	"net/http"
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

func (engine *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	ctx := newContext(w, req)
	ctx.handlers = engine.RouterGroup.middlewares
	engine.router.handle(ctx)
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
		router:      nil,
		RouterGroup: RouterGroup{
			Handlers: nil,
			basePath: "/",
			root: true,
		},
		ForwardedByClientIP:true,
		groups:      nil,
	}




	engine.RouterGroup.engine = engine



	return engine
}

func Default() *Engine {
	engine := New()
	engine.Use(Logger(), Recovery())
	return engine
}
