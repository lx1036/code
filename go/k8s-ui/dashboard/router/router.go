package router

import (
	"github.com/gin-gonic/gin"
	"k8s-lx1036/k8s-ui/dashboard/controllers/auth/authenticator"
	"k8s-lx1036/k8s-ui/dashboard/controllers/auth/csrf"
	"k8s-lx1036/k8s-ui/dashboard/controllers/resource/configmap"
	"k8s-lx1036/k8s-ui/dashboard/controllers/resource/deployment"
	"k8s-lx1036/k8s-ui/dashboard/controllers/resource/event"
	"k8s-lx1036/k8s-ui/dashboard/controllers/resource/namespace"
	"k8s-lx1036/k8s-ui/dashboard/controllers/resource/node"
	"k8s-lx1036/k8s-ui/dashboard/controllers/resource/persistentvolume"
	"k8s-lx1036/k8s-ui/dashboard/controllers/resource/persistentvolumeclaim"
	"k8s-lx1036/k8s-ui/dashboard/controllers/resource/pod"
	"k8s-lx1036/k8s-ui/dashboard/controllers/resource/role"
	"k8s-lx1036/k8s-ui/dashboard/controllers/resource/rolebinding"
	"k8s-lx1036/k8s-ui/dashboard/controllers/resource/secret"
)

func SetupRouter() *gin.Engine {
	router := gin.Default()
	api := router.Group("/api/v1")
	{
		// csrftoken
		api.GET("/csrftoken/:action", (&csrf.CsrfController{}).GetCsrfToken())

		// login
		api.GET("/login/modes", (&authenticator.AuthenticationController{}).GetLoginModes())
		api.GET("/login/skippable", (&authenticator.AuthenticationController{}).GetLoginSkippable())

		/* Core Concepts */
		// namespace
		api.GET("/namespaces", (&namespace.NamespaceController{}).ListNamespaces())
		api.GET("/namespaces/{namespace}", (&namespace.NamespaceController{}).GetNamespace())
		api.POST("/namespace", (&namespace.NamespaceController{}).CreateNamespaces())
		// deployment
		api.POST("/appdeployment", (&deployment.DeploymentController{}).HandleDeploy())
		api.POST("/appdeployment/validate/name", (&deployment.DeploymentController{}).HandleNameValidity())
		// pod
		api.GET("/pod", (&pod.PodController{}).ListPods())
		api.GET("/pod/{namespace}", (&pod.PodController{}).ListPods())
		api.GET("/pod/{namespace}/{pod}", (&pod.PodController{}).GetPod())
		api.GET("/pod/{namespace}/{pod}/containers", (&pod.PodController{}).ListPodContainers())
		api.GET("/pod/{namespace}/{pod}/shell/{container}", (&pod.PodController{}).ExecPodShell())
		api.GET("/pod/{namespace}/{pod}/persistentvolumeclaim", (&pod.PodController{}).GetPodPvc())

		/* Cluster */
		// node
		api.GET("/node", (&node.NodeController{}).ListNodes())
		api.GET("/node/{node}", (&node.NodeController{}).GetNode())
		api.GET("/node/{node}/pods", (&node.NodeController{}).ListNodePods())

		/* Controller */
		// replication controller
		api.GET("/replicationcontroller")
		//api.GET("/replicationcontroller/:namespace/:replicationController/event")

		/* Storage */
		// secret
		api.GET("/secrets", (&secret.SecretController{}).ListSecrets())
		api.GET("/secrets/{namespace}", (&secret.SecretController{}).ListSecrets())
		api.GET("/secrets/{namespace}/{secret}", (&secret.SecretController{}).GetSecret())
		api.POST("/secret", (&secret.SecretController{}).CreateSecret())
		// configmap
		api.GET("/configmaps", (&configmap.ConfigmapController{}).ListConfigmaps())
		api.GET("/configmaps/{namespace}", (&configmap.ConfigmapController{}).ListConfigmaps())
		api.GET("/configmaps/{namespace}/{configmap}", (&configmap.ConfigmapController{}).GetConfigmap())
		api.POST("/configmap", (&configmap.ConfigmapController{}).CreateConfigmap())
		// pv
		api.GET("/persistentvolumes", (&persistentvolume.PersistentVolumeController{}).ListPersistentVolumes())
		api.GET("/persistentvolumes/{persistentvolume}", (&persistentvolume.PersistentVolumeController{}).GetPersistentVolume())
		api.GET("/persistentvolumes/namespace/{namespace}/name/{persistentvolume}", (&persistentvolume.PersistentVolumeController{}).GetPersistentVolume())
		// pvc
		api.GET("/persistentvolumeclaims", (&persistentvolumeclaim.PersistentVolumeClaimController{}).ListPersistentVolumeClaims())
		api.GET("/persistentvolumeclaims/{persistentvolumeclaim}", (&persistentvolumeclaim.PersistentVolumeClaimController{}).GetPersistentVolumeClaim())
		api.GET("/persistentvolumeclaims/{namespace}/name", (&persistentvolumeclaim.PersistentVolumeClaimController{}).GetPersistentVolumeClaim())

		/* Network */

		/* Security */
		// role
		api.GET("/role/{namespace}", (&role.RoleController{}).ListRoles())
		api.GET("/role/{namespace}/{role}", (&role.RoleController{}).GetRole())
		// rolebinding
		api.GET("/rolebinding/{namespace}", (&rolebinding.RoleBindingController{}).ListRoleBindings())
		api.GET("/rolebinding/{namespace}/{rolebinding}", (&rolebinding.RoleBindingController{}).GetRoleBinding())

		/* Log/Monitor */
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

		/* CRD */
	}

	return router
}
