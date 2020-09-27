package routers

import (
	"github.com/gin-gonic/gin"
	cors "github.com/rs/cors/wrapper/gin"
	"k8s-lx1036/k8s-ui/backend/controllers"
	"k8s-lx1036/k8s-ui/backend/controllers/auth"
	"k8s-lx1036/k8s-ui/backend/controllers/kubernetes"
	"k8s-lx1036/k8s-ui/backend/controllers/kubernetes/service"
	"k8s-lx1036/k8s-ui/backend/routers/middlewares"
)

func SetupRouter() *gin.Engine {
	router := gin.Default()

	// application's global HTTP middleware stack
	router.Use(cors.AllowAll()) // cors

	router.POST("/login/:type", (&auth.AuthController{}).Login())
	router.GET("/logout", (&auth.AuthController{}).Logout())
	router.GET("/api/v1/configs/base", (&controllers.BaseConfigController{}).ListBase())

	authorizedRouter := router.Group("/")
	authorizedRouter.Use(middlewares.AuthRequired())
	authorizedRouter.GET(`/me`, (&auth.AuthController{}).CurrentUser())
	apiV1Router := authorizedRouter.Group("/api/v1")

	{
		apiV1Router.GET("/notifications/subscribe", (&controllers.NotificationController{}).Subscribe())
		apiV1Router.POST("/notifications", (&controllers.NotificationController{}).Create())
		apiV1Router.GET("/notifications", (&controllers.NotificationController{}).List())
		apiV1Router.GET("/apps/statistics", (&controllers.AppController{}).AppStatistics())
		apiV1Router.GET("/users/statistics", (&controllers.UserController{}).UserStatistics())
		apiV1Router.GET("/kubernetes/pods/statistics", (&kubernetes.KubePodController{}).PodStatistics())
		apiV1Router.GET("/kubernetes/nodes/statistics", (&kubernetes.KubeNodeController{}).PodStatistics())

		// edge-nodes
		apiV1Router.GET("/services/edgenodes", (&service.EdgeNodeController{}).List())
		apiV1Router.DELETE("/apps/:appId/services/:serviceId", (&service.ServiceController{}).Delete())
	}

	{
		apiV1Router.GET("/apps/:appId/_proxy/clusters/:cluster/namespaces/:namespace/:kind/:kindName")
	}

	return router
}
