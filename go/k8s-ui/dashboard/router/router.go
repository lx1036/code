package router

import (
	"github.com/gin-gonic/gin"
	"k8s-lx1036/k8s-ui/dashboard/controllers/auth/authenticator"
	"k8s-lx1036/k8s-ui/dashboard/controllers/auth/csrf"
	"k8s-lx1036/k8s-ui/dashboard/controllers/plugin"
	"k8s-lx1036/k8s-ui/dashboard/controllers/resource/deployment"
	"k8s-lx1036/k8s-ui/dashboard/controllers/resource/event"
	"k8s-lx1036/k8s-ui/dashboard/controllers/resource/pod"
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

		// deployment
		api.POST("/appdeployment", (&deployment.DeploymentController{}).HandleDeploy())
		api.POST("/appdeployment/validate/name", (&deployment.DeploymentController{}).HandleNameValidity())

		// pod
		api.GET("/pod", (&pod.PodController{}).ListNamespacePod())
		api.GET("/pod/{namespace}", (&pod.PodController{}).ListNamespacePod())
		api.GET("/pod/{namespace}/{pod}", (&pod.PodController{}).ListNamespacePod())
		api.GET("/pod/{namespace}/{pod}/container", (&pod.PodController{}).ListNamespacePod())
		api.GET("/pod/{namespace}/{pod}/shell/{container}", (&pod.PodController{}).ListNamespacePod())
		api.GET("/pod/{namespace}/{pod}/persistentvolumeclaim", (&pod.PodController{}).ListNamespacePod())
		
		// event
		api.GET("/namespace/{namespace}/event", (&event.EventController{}).ListNamespaceEvents())
		api.GET("/pod/{namespace}/{pod}/event", (&event.EventController{}).ListPodEvents())

		// Replication
		api.GET("/replicationcontroller")
		//api.GET("/replicationcontroller/:namespace/:replicationController/event")
	}

	return router
}
