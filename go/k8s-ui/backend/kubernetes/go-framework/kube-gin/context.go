package kube_gin

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
)

type Context struct {
	Writer     ResponseWriter
	Req        *http.Request
	Request   *http.Request
	Path       string
	Method     string
	StatusCode int
	Params     map[string]string // 动态路由参数

	handlers []HandlerFunc
	index    int

	// Keys is a key/value pair exclusively for the context of each request.
	Keys map[string]interface{}
	engine *Engine
	// Errors is a list of errors attached to all the handlers/middlewares who used this context.
	Errors errorMsgs
}

type H map[string]interface{}

//func newContext(writer http.ResponseWriter, request *http.Request) *Context {
//	return &Context{
//		Writer: writer,
//		Req:    request,
//		Path:   request.URL.Path,
//		Method: request.Method,
//	}
//}

func (context *Context) Next() {
	context.index++
	for ; context.index < len(context.handlers); context.index++ {
		context.handlers[context.index](context)
	}
}

func (context *Context) Query(key string) string {
	return context.Req.URL.Query().Get(key)
}

func (context *Context) PostForm(key string) string {
	return context.Req.FormValue(key)
}

func (context *Context) SetHeader(key string, value string) {
	context.Writer.Header().Set(key, value)
}

func (context *Context) AddHeader(key string, value string) {
	context.Writer.Header().Add(key, value)
}

func (context *Context) Status(code int) {
	context.StatusCode = code
	context.Writer.WriteHeader(code)
}

func (context *Context) HTML(code int, html string) {
	context.SetHeader("Content-Type", "text/html")
	context.Status(code)
	_, _ = context.Writer.Write([]byte(html))
}

func (context *Context) String(code int, format string, values ...interface{}) {
	context.SetHeader("Content-Type", "text/plain")
	context.Status(code)
	_, _ = context.Writer.Write([]byte(fmt.Sprintf(format, values...)))
}

func (context *Context) JSON(code int, obj interface{}) {
	context.SetHeader("Content-Type", "application/json")
	context.Status(code)
	if err := json.NewEncoder(context.Writer).Encode(obj); err != nil {
		http.Error(context.Writer, err.Error(), http.StatusInternalServerError)
	}
}

func (context *Context) Fail(code int, err string) {
	context.index = len(context.handlers)
	context.JSON(code, H{
		"errno":    -1,
		"errormsg": err,
	})
}

func (context *Context) ClientIP() string {
	if context.engine.ForwardedByClientIP {
		clientIP := context.requestHeader("X-Forwarded-For")
		clientIP = strings.TrimSpace(strings.Split(clientIP, ",")[0])
		if clientIP == "" {
			clientIP = strings.TrimSpace(context.requestHeader("X-Real-Ip"))
		}
		if clientIP != "" {
			return clientIP
		}
	}

	if ip, _, err := net.SplitHostPort(strings.TrimSpace(context.Request.RemoteAddr)); err == nil {
		return ip
	}

	return ""
}

func (context *Context) requestHeader(key string) string {
	return context.Request.Header.Get(key)
}
