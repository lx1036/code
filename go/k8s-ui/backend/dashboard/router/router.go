package router

import (
	"github.com/gin-gonic/gin"
	"k8s-lx1036/k8s-ui/backend/dashboard/controllers/resource/deployment"
)

func SetupRouter() *gin.Engine {
	router := gin.Default()
	api := router.Group("/api")
	{
		// Deployment
		api.POST("/appdeployment", (&deployment.DeploymentController{}).HandleDeploy())
		api.POST("/appdeployment", (&deployment.DeploymentController{}).HandleNameValidity())

		// Replication
		api.GET("/replicationcontroller", )
		//api.GET("/replicationcontroller/:namespace/:replicationController/event")
	}

	return router
}
