package event

import (
	"github.com/gin-gonic/gin"
	"k8s-lx1036/k8s-ui/dashboard/controllers/resource/common"
	"k8s-lx1036/k8s-ui/dashboard/controllers/resource/common/client"
	"k8s-lx1036/k8s-ui/dashboard/controllers/resource/common/dataselect"
	"net/http"
)

type EventController struct {
}

func (controller *EventController) ListNamespaceEvents() gin.HandlerFunc {
	return func(context *gin.Context) {
		k8sClient := client.DefaultClientManager.Client()
		namespaceName := context.Param("namespace")
		dataSelectQuery := dataselect.ParseDataSelectFromRequest(context)
		events, _ := ListNamespaceEventsByQuery(k8sClient, namespaceName, dataSelectQuery)

		context.JSON(http.StatusOK, common.JsonResponse{
			Errno:  0,
			Errmsg: "success",
			Data:   events,
		})
	}
}

func (controller *EventController) ListPodEvents() gin.HandlerFunc {
	return func(context *gin.Context) {
		k8sClient := client.DefaultClientManager.Client()
		namespaceName := context.Param("namespace")
		podName := context.Param("pod")
		dataSelectQuery := dataselect.ParseDataSelectFromRequest(context)
		events, _ := ListResourceEventsByQuery(k8sClient, namespaceName, podName, dataSelectQuery)

		context.JSON(http.StatusOK, common.JsonResponse{
			Errno:  0,
			Errmsg: "success",
			Data:   events,
		})
	}
}

func (controller *EventController) ListReplicationControllerEvents() gin.HandlerFunc {
	return func(context *gin.Context) {
		k8sClient := client.DefaultClientManager.Client()
		namespaceName := context.Param("namespace")
		replicationControllerName := context.Param("replicationcontroller")
		dataSelectQuery := dataselect.ParseDataSelectFromRequest(context)
		events, _ := ListResourceEventsByQuery(k8sClient, namespaceName, replicationControllerName, dataSelectQuery)

		context.JSON(http.StatusOK, common.JsonResponse{
			Errno:  0,
			Errmsg: "success",
			Data:   events,
		})
	}
}

func (controller *EventController) ListReplicationSetEvents() gin.HandlerFunc {
	return func(context *gin.Context) {
		k8sClient := client.DefaultClientManager.Client()
		namespaceName := context.Param("namespace")
		replicaSetName := context.Param("replicaset")
		dataSelectQuery := dataselect.ParseDataSelectFromRequest(context)
		events, _ := ListResourceEventsByQuery(k8sClient, namespaceName, replicaSetName, dataSelectQuery)

		context.JSON(http.StatusOK, common.JsonResponse{
			Errno:  0,
			Errmsg: "success",
			Data:   events,
		})
	}
}

func (controller *EventController) ListDeploymentEvents() gin.HandlerFunc {
	return func(context *gin.Context) {
		k8sClient := client.DefaultClientManager.Client()
		namespaceName := context.Param("namespace")
		deploymentName := context.Param("deployment")
		dataSelectQuery := dataselect.ParseDataSelectFromRequest(context)
		events, _ := ListResourceEventsByQuery(k8sClient, namespaceName, deploymentName, dataSelectQuery)

		context.JSON(http.StatusOK, common.JsonResponse{
			Errno:  0,
			Errmsg: "success",
			Data:   events,
		})
	}
}

func (controller *EventController) ListDeamonSetEvents() gin.HandlerFunc {
	return func(context *gin.Context) {
		k8sClient := client.DefaultClientManager.Client()
		namespaceName := context.Param("namespace")
		deamonSetName := context.Param("deamonset")
		dataSelectQuery := dataselect.ParseDataSelectFromRequest(context)
		events, _ := ListResourceEventsByQuery(k8sClient, namespaceName, deamonSetName, dataSelectQuery)

		context.JSON(http.StatusOK, common.JsonResponse{
			Errno:  0,
			Errmsg: "success",
			Data:   events,
		})
	}
}

func (controller *EventController) ListJobEvents() gin.HandlerFunc {
	return func(context *gin.Context) {
		k8sClient := client.DefaultClientManager.Client()
		namespaceName := context.Param("namespace")
		jobName := context.Param("job")
		dataSelectQuery := dataselect.ParseDataSelectFromRequest(context)
		events, _ := ListResourceEventsByQuery(k8sClient, namespaceName, jobName, dataSelectQuery)

		context.JSON(http.StatusOK, common.JsonResponse{
			Errno:  0,
			Errmsg: "success",
			Data:   events,
		})
	}
}

func (controller *EventController) ListCronjobEvents() gin.HandlerFunc {
	return func(context *gin.Context) {
		k8sClient := client.DefaultClientManager.Client()
		namespaceName := context.Param("namespace")
		cronjobName := context.Param("cronjob")
		dataSelectQuery := dataselect.ParseDataSelectFromRequest(context)
		events, _ := ListResourceEventsByQuery(k8sClient, namespaceName, cronjobName, dataSelectQuery)

		context.JSON(http.StatusOK, common.JsonResponse{
			Errno:  0,
			Errmsg: "success",
			Data:   events,
		})
	}
}

func (controller *EventController) ListServiceEvents() gin.HandlerFunc {
	return func(context *gin.Context) {
		k8sClient := client.DefaultClientManager.Client()
		namespaceName := context.Param("namespace")
		serviceName := context.Param("service")
		dataSelectQuery := dataselect.ParseDataSelectFromRequest(context)
		events, _ := ListResourceEventsByQuery(k8sClient, namespaceName, serviceName, dataSelectQuery)

		context.JSON(http.StatusOK, common.JsonResponse{
			Errno:  0,
			Errmsg: "success",
			Data:   events,
		})
	}
}

func (controller *EventController) ListStatefulSetEvents() gin.HandlerFunc {
	return func(context *gin.Context) {
		k8sClient := client.DefaultClientManager.Client()
		namespaceName := context.Param("namespace")
		statefulSetName := context.Param("statefulset")
		dataSelectQuery := dataselect.ParseDataSelectFromRequest(context)
		events, _ := ListResourceEventsByQuery(k8sClient, namespaceName, statefulSetName, dataSelectQuery)

		context.JSON(http.StatusOK, common.JsonResponse{
			Errno:  0,
			Errmsg: "success",
			Data:   events,
		})
	}
}

func (controller *EventController) ListNodeEvents() gin.HandlerFunc {
	return func(context *gin.Context) {
		k8sClient := client.DefaultClientManager.Client()
		namespaceName := context.Param("namespace")
		nodeName := context.Param("node")
		dataSelectQuery := dataselect.ParseDataSelectFromRequest(context)
		events, _ := ListResourceEventsByQuery(k8sClient, namespaceName, nodeName, dataSelectQuery)

		context.JSON(http.StatusOK, common.JsonResponse{
			Errno:  0,
			Errmsg: "success",
			Data:   events,
		})
	}
}

func (controller *EventController) ListCrdEvents() gin.HandlerFunc {
	return func(context *gin.Context) {
		k8sClient := client.DefaultClientManager.Client()
		namespaceName := context.Param("namespace")
		objectName := context.Param("object")
		dataSelectQuery := dataselect.ParseDataSelectFromRequest(context)
		events, _ := ListResourceEventsByQuery(k8sClient, namespaceName, objectName, dataSelectQuery)

		context.JSON(http.StatusOK, common.JsonResponse{
			Errno:  0,
			Errmsg: "success",
			Data:   events,
		})
	}
}
