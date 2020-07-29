package middlewares

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

const (
	MaxAge = "86400"
)

func Cors() gin.HandlerFunc {
	return func(context *gin.Context) {
		context.Next()
		if context.Request.Method == http.MethodOptions {
			method := context.Request.Header.Get("Access-Control-Request-Method")
			if method != "" {
				context.Writer.Header().Set("Access-Control-Allow-Origin", "*")
				context.Writer.Header().Set("Access-Control-Allow-Methods", "*")
				context.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
				context.Writer.Header().Set("Access-Control-Expose-Headers", "*")
				context.Writer.Header().Set("Access-Control-Max-Age", MaxAge)
			}

			if !context.Writer.Written() {
				context.Writer.WriteHeader(http.StatusNoContent)
			}
		} else {
			// https://www.w3.org/TR/cors/#resource-requests
			context.Writer.Header().Set("Access-Control-Allow-Origin", "*")
			context.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
			context.Writer.Header().Set("Access-Control-Expose-Headers", "*")
		}
	}
}
