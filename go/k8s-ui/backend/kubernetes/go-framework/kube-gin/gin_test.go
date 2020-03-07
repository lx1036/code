package kube_gin

import (
	"net/http"
	"testing"
)

func TestEngine_Step1(test *testing.T) {
	/*engine := New()
	engine.Get("/", func(writer http.ResponseWriter, request *http.Request) {
		_, _ = fmt.Fprintf(writer, "ok\n")
	})

	_ = engine.Run(":9999")*/
}

func TestEngine_Step2(test *testing.T) {
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
