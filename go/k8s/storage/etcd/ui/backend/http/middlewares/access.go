package middlewares

import (
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"net/url"
	"time"
)

// 记录access日志
func AccessLog() gin.HandlerFunc {
	return func(context *gin.Context) {
		start := time.Now()

		context.Next()

		latency := time.Since(start).Milliseconds()
		// access the status we are sending
		status := context.Writer.Status()
		requestUrl, _ := url.QueryUnescape(context.Request.URL.String())
		log.WithFields(log.Fields{
			"method":  context.Request.Method,
			"url":     requestUrl,
			"status":  status,
			"latency": latency,
		}).Info("REQUEST INFO")
	}
}
