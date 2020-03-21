package router

import (
	"github.com/gin-gonic/gin"
	"k8s-lx1036/k8s-ui/backend/dashboard/controllers/auth/authenticator"
	"k8s-lx1036/k8s-ui/backend/dashboard/controllers/auth/csrf"
	"k8s-lx1036/k8s-ui/backend/dashboard/controllers/plugin"
	"k8s-lx1036/k8s-ui/backend/dashboard/controllers/resource/deployment"
	"k8s-lx1036/k8s-ui/backend/dashboard/controllers/resource/pod"
)

func SetupRouter() *gin.Engine {
	router := gin.Default()
	api := router.Group("/api/v1")
	{
		// csrftoken
		api.GET("/csrftoken/:action", (&csrf.CsrfController{}).GetCsrfToken())

		// plugin
		api.GET("/plugin/config", (&plugin.PluginController{}).HandleConfig())

		// login
		api.GET("/login/modes", (&authenticator.AuthenticationController{}).GetLoginModes())
		api.GET("/login/skippable", (&authenticator.AuthenticationController{}).GetLoginSkippable())

		// Deployment
		api.POST("/appdeployment", (&deployment.DeploymentController{}).HandleDeploy())
		api.POST("/appdeployment/validate/name", (&deployment.DeploymentController{}).HandleNameValidity())

		// Pod
		api.GET("/pod", (&pod.PodController{}).List())

		// Replication
		api.GET("/replicationcontroller")
		//api.GET("/replicationcontroller/:namespace/:replicationController/event")
	}

	return router
}
