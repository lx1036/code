package router

import (
	"github.com/gin-gonic/gin"
	"k8s-lx1036/k8s-ui/backend/dashboard/controllers/resource/deployment"
)

func SetupRouter() *gin.Engine {
	router := gin.Default()
	api := router.Group("/api")
	{
		api.POST("/appdeployment", (&deployment.DeploymentController{}).HandleDeploy())
	}

	return router
}
