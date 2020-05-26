package event

import (
	"github.com/gin-gonic/gin"
	"k8s-lx1036/k8s-ui/dashboard/client"
	"k8s-lx1036/k8s-ui/dashboard/controllers/resource/common/dataselect"
	"net/http"
)

type EventController struct {
}

func (controller *EventController)  ListNamespaceEvents() gin.HandlerFunc {
	return func(context *gin.Context) {
		k8sClient := client.DefaultClientManager.Client()
		namespaceName := context.Param("namespace")
		dataSelectQuery := dataselect.ParseDataSelectFromRequest(context)
		events, _ := ListNamespaceEventsByQuery(k8sClient, namespaceName, dataSelectQuery)
		
		return context.JSON(http.StatusOK, gin.H{
		
		})
	}
}
