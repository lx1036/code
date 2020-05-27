package router

import (
	"github.com/gin-gonic/gin"
	"k8s-lx1036/k8s-ui/dashboard/controllers/auth/authenticator"
	"k8s-lx1036/k8s-ui/dashboard/controllers/auth/csrf"
	"k8s-lx1036/k8s-ui/dashboard/controllers/resource/deployment"
	"k8s-lx1036/k8s-ui/dashboard/controllers/resource/event"
	"k8s-lx1036/k8s-ui/dashboard/controllers/resource/namespace"
	"k8s-lx1036/k8s-ui/dashboard/controllers/resource/pod"
)

func SetupRouter() *gin.Engine {
	router := gin.Default()
	api := router.Group("/api/v1")
	{
		// csrftoken
		api.GET("/csrftoken/:action", (&csrf.CsrfController{}).GetCsrfToken())

		
		// namespace
		api.GET("/namespaces", (&namespace.NamespaceController{}).ListNamespaces())
		api.GET("/namespaces/{namespace}", (&namespace.NamespaceController{}).GetNamespace())
		api.POST("/namespace", (&namespace.NamespaceController{}).CreateNamespaces())

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
		api.GET("/replicationcontroller/{namespace}/{replicationcontroller}/event", (&event.EventController{}).ListReplicationControllerEvents())
		api.GET("/replicaset/{namespace}/{replicaset}/event", (&event.EventController{}).ListReplicationSetEvents())
		api.GET("/deployment/{namespace}/{deployment}/event", (&event.EventController{}).ListDeploymentEvents())
		api.GET("/deamonset/{namespace}/{deamonset}/event", (&event.EventController{}).ListDeamonSetEvents())
		api.GET("/job/{namespace}/{job}/event", (&event.EventController{}).ListJobEvents())
		api.GET("/cronjob/{namespace}/{cronjob}/event", (&event.EventController{}).ListCronjobEvents())
		api.GET("/service/{namespace}/{service}/event", (&event.EventController{}).ListServiceEvents())
		api.GET("/statefulset/{namespace}/{statefulset}/event", (&event.EventController{}).ListStatefulSetEvents())
		api.GET("/node/{namespace}/{node}/event", (&event.EventController{}).ListNodeEvents())
		api.GET("/crd/{namespace}/{crd}/{object}/event", (&event.EventController{}).ListCrdEvents())

		// replication controller
		api.GET("/replicationcontroller")
		//api.GET("/replicationcontroller/:namespace/:replicationController/event")
	}

	return router
}
