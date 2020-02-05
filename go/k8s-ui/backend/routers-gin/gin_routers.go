package routers_gin

import (
	"github.com/gin-gonic/gin"
	cors "github.com/rs/cors/wrapper/gin"
	"k8s-lx1036/k8s-ui/backend/controllers/auth"
)

func SetupRouter() *gin.Engine {
	router := gin.Default()

	// application's global HTTP middleware stack
	router.Use(cors.Default()) // cors

	controllerRegistry := New(router)

	// Auth
	controllerRegistry.AddRouter("GET", `/login/:type`, &auth.AuthController{}, "Login")
	controllerRegistry.AddRouter("POST", `/login/:type`, &auth.AuthController{}, "Login")
	controllerRegistry.AddRouter("GET", `/logout`, &auth.AuthController{}, "Logout")
	controllerRegistry.AddRouter("GET", `/me`, &auth.AuthController{}, "CurrentUser")

	return router
}