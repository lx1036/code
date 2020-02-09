package routers_gin

import (
	"github.com/gin-gonic/gin"
	cors "github.com/rs/cors/wrapper/gin"
	"k8s-lx1036/k8s-ui/backend/controllers"
	"k8s-lx1036/k8s-ui/backend/controllers/auth"
	"k8s-lx1036/k8s-ui/backend/routers-gin/middlewares"
)

func SetupRouter() *gin.Engine {
	router := gin.Default()

	// application's global HTTP middleware stack
	router.Use(cors.AllowAll()) // cors

	router.POST("/login/:type", (&auth.AuthController{}).Login())
	router.GET("/logout", (&auth.AuthController{}).Logout())

	authorizedRouter := router.Group("/")
	authorizedRouter.Use(middlewares.AuthRequired())
	{
		authorizedRouter.GET(`/me`, (&auth.AuthController{}).CurrentUser())
		apiV1Router := authorizedRouter.Group("/api/v1")
		apiV1Router.GET("/configs/base", (&controllers.BaseConfigController{}).ListBase())

		apiV1Router.GET("/notifications/subscribe", (&controllers.NotificationController{}).Subscribe())
		apiV1Router.POST("/notifications", (&controllers.NotificationController{}).Create())
		apiV1Router.GET("/notifications", (&controllers.NotificationController{}).List())
	}

	return router
}
