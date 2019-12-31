package cors

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type corsWrapper struct {
	*Cors
	OptionPassthrough bool
}

func (cors corsWrapper) build() gin.HandlerFunc {
	return func(context *gin.Context) {
		cors.HandlerFunc(context.Writer, context.Request)
		if !cors.OptionPassthrough &&
			context.Request.Method == http.MethodOptions &&
			context.GetHeader("Access-Control-Request-Method") != "" {
			// Abort processing next Gin middlewares.
			context.AbortWithStatus(http.StatusOK)
		}
	}
}

func CorsMiddleware() gin.HandlerFunc {
	return corsWrapper{Cors: Default()}.build()
}

func WrapperAllowAll() gin.HandlerFunc {
	return corsWrapper{Cors: AllowAll()}.build()
}
