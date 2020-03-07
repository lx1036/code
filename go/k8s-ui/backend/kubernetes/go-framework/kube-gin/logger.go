package kube_gin

import (
	"log"
	"time"
)

func Logger() HandlerFunc {
	return func(context *Context) {
		now := time.Now()
		context.Next()
		log.Printf("[%d] %s %v", context.StatusCode, context.Req.RequestURI, time.Since(now).Milliseconds())
	}
}
