package middlewares

import (
	"app/http/log/prometheus"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"time"
)

// Prometheus access 监控
func PrometheusAccessLogger() gin.HandlerFunc {
	return func(context *gin.Context) {
		start := time.Now()

		context.Next()

		go func() {
			if prometheus.GetWrapper() != nil {
				// 路径限制白名单
				if _, ok := prometheus.GetWrapper().WatchPath[context.Request.URL.Path]; ok {
					_, _ = prometheus.GetWrapper().QpsCountLog(prometheus.QPSRecord{
						Times:  float64(1),
						Api:    context.Request.URL.Path,
						Module: prometheus.Module,
						Method: context.Request.Method,
						Code:   context.Writer.Status(),
					})

					latency := float64(time.Since(start).Milliseconds())
					_, _ = prometheus.GetWrapper().LatencyLog(prometheus.LatencyRecord{
						Time:   latency,
						Api:    context.Request.URL.Path,
						Module: prometheus.Module,
						Method: context.Request.Method,
					})

					log.WithFields(log.Fields{
						"method":  context.Request.Method,
						"path":    context.Request.URL.Path,
						"status":  context.Writer.Status(),
						"latency": latency,
					}).Info("[app level]prometheus access logger")
				}
			} else {
				log.Warn("need to init prometheus!!!")
			}
		}()
	}
}
