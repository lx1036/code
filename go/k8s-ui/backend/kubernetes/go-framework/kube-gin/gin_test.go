package kube_gin

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"reflect"
	"testing"
)

func TestEngine_Step1(test *testing.T) {
	/*engine := New()
	engine.Get("/", func(writer http.ResponseWriter, request *http.Request) {
		_, _ = fmt.Fprintf(writer, "ok\n")
	})

	_ = engine.Run(":9999")*/
}

func TestContext_Step2(test *testing.T) {
	engine := New()
	engine.Get("/", func(context *Context) {
		context.HTML(http.StatusOK, "<h1>hello world</h1>")
	})

	engine.Get("/hello", func(context *Context) {
		context.String(http.StatusOK, "hello %s,the url path is %s", context.Query("name"), context.Path)
	})

	engine.Post("/login", func(context *Context) {
		context.JSON(http.StatusOK, H{
			"username": context.PostForm("username"),
			"password": context.PostForm("password"),
		})
	})

	_ = engine.Run(":9999")
}

func TestParsePattern(test *testing.T) {
	ok := reflect.DeepEqual(parsePattern("/p/:name"), []string{"p", ":name"})
	ok = ok && reflect.DeepEqual(parsePattern("/p/*"), []string{"p", "*"})
	ok = ok && reflect.DeepEqual(parsePattern("/p/*name/*"), []string{"p", "*name"})
	if !ok {
		test.Fatal("test parsePattern failed")
	}
}

func TestDynamicRoutes_Step3(test *testing.T) {
	engine := New()
	engine.Get("/", nil)
	engine.Get("/people/*name/accounts", nil)
	engine.Get("/people/:id/accounts", nil)

	accountsNode, params := engine.router.getRoute("GET", "/people/lx1036/accounts")
	assert.Equal(test, "lx1036/accounts", params["name"])
	assert.Equal(test, "/people/*name/accounts", accountsNode.pattern)

	engine.Get("/people/:id", nil)
	peopleNode, params := engine.router.getRoute("GET", "/people/lx1036")
	assert.Equal(test, "lx1036", params["id"])
	assert.Equal(test, "/people/:id", peopleNode.pattern)
}

func TestDynamicRoutes_Engine_Step3(test *testing.T) {
	engine := New()
	engine.Get("/people/:id", func(context *Context) {
		id := context.Params["id"]
		context.JSON(http.StatusOK, H{
			"errno":  0,
			"errmsg": "success",
			"data":   id,
		})
	})

	_ = engine.Run(":9999")
}

func TestGroupRoutes_Step4(test *testing.T) {
	engine := New()
	v1 := engine.Group("/v1")
	{
		v1.Get("/people/:id", func(context *Context) {
			id := context.Params["id"]
			context.JSON(http.StatusOK, H{
				"errno":  0,
				"errmsg": "success ",
				"data":   fmt.Sprintf("%s %s", "/v1", id),
			})
		})
	}

	v2 := engine.Group("/v2")
	{
		v2.Get("/people/:id", func(context *Context) {
			id := context.Params["id"]
			context.JSON(http.StatusOK, H{
				"errno":  0,
				"errmsg": "success",
				"data":   fmt.Sprintf("%s %s", "/v2", id),
			})
		})
	}

	v2Alpha := v2.Group("/alpha")
	{
		v2Alpha.Get("/people/:id", func(context *Context) {
			id := context.Params["id"]
			context.JSON(http.StatusOK, H{
				"errno":  0,
				"errmsg": "success",
				"data":   fmt.Sprintf("%s%s %s", "/v2", "/alpha", id),
			})
		})
	}

	_ = engine.Run(":9999")
}
