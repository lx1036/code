package gin

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"testing"
)

func TestGinMiddleware(test *testing.T) {
	engine := gin.Default()
	engine.GET("/hello", func(context *gin.Context) {
		context.String(http.StatusOK, "hello %s,the url path is %s", context.Query("name"), context.Request.URL.Path)
	})
	_ = engine.Run(":9999")
}
