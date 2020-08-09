package router

import (
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"k8s-lx1036/k8s/storage/etcd/ui/backend/http/controllers"
	"k8s-lx1036/k8s/storage/etcd/ui/backend/http/middlewares"
	"k8s-lx1036/k8s/storage/etcd/ui/backend/http/prometheus"
	"os"
)

func SetupRouter() *gin.Engine {
	router := gin.Default()
	router.Use(middlewares.Cors())

	// 接入Prometheus
	ProjectName := os.Getenv("PROJECT_NAME")
	Idc := "dev" // Idc : 机房标识
	if ProjectName != "" || viper.GetBool("debug") {
		if viper.GetBool("debug") {
			if ProjectName == "" {
				ProjectName = "demo_project"
			}
		}
		prometheus.Init(prometheus.Opts{
			Idc:             Idc,
			AppName:         ProjectName,
			HistogramBucket: []float64{100, 124, 154, 191, 237, 295, 367, 456, 567, 705, 876, 1089, 1354, 1683, 2092, 2601, 3234, 4021, 5000, 5001},
			WatchPath: map[string]struct{}{
				"/api/v1/member": {},
			},
		})
		router.Use(middlewares.PrometheusAccessLogger())
		prometheus.MetricsServerStart("/metrics", 18081)
	}

	router.Use(middlewares.AccessLog())
	routerV1 := router.Group("/api/v1")
	routerV1.GET("/members", (&controllers.EtcdController{}).ListMembers())
	routerV1.GET("/servers", (&controllers.EtcdController{}).ListServers())
	routerV1.POST("/key", (&controllers.EtcdController{}).CreateKey())
	routerV1.GET("/list", (&controllers.EtcdController{}).List())
	routerV1.GET("/key", (&controllers.EtcdController{}).GetKey())
	routerV1.PUT("/key", (&controllers.EtcdController{}).UpdateKey())
	routerV1.DELETE("/key", (&controllers.EtcdController{}).DeleteKey())
	routerV1.GET("/key/format", (&controllers.EtcdController{}).GetKeyFormat())
	routerV1.GET("/logs", (&controllers.EtcdController{}).GetLogs())
	routerV1.GET("/users", (&controllers.EtcdController{}).GetUsers())
	routerV1.GET("/logtypes", (&controllers.EtcdController{}).GetLogTypes())

	return router
}
