package node

import (
	"github.com/gin-gonic/gin"
	"k8s-lx1036/k8s-ui/dashboard/controllers/resource/common"
	"k8s-lx1036/k8s-ui/dashboard/controllers/resource/common/client"
	"k8s-lx1036/k8s-ui/dashboard/controllers/resource/common/dataselect"
	"net/http"
)

type NodeController struct {
}

func (controller *NodeController) ListNodes() gin.HandlerFunc {
	return func(context *gin.Context) {
		k8sClient := client.DefaultClientManager.Client()
		dataSelectQuery := dataselect.ParseDataSelectFromRequest(context)

		nodes, _ := ListNodesByQuery(k8sClient, dataSelectQuery)

		context.JSON(http.StatusOK, common.JsonResponse{
			Errno:  0,
			Errmsg: "success",
			Data:   events,
		})
	}
}

func (controller *NodeController) GetNode() gin.HandlerFunc {
	return func(context *gin.Context) {

	}
}

func (controller *NodeController) ListNodePods() gin.HandlerFunc {
	return func(context *gin.Context) {

	}
}
