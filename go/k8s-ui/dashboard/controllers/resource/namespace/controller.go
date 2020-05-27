package namespace

import (
	"github.com/gin-gonic/gin"
	"k8s-lx1036/k8s-ui/dashboard/controllers/resource/common"
	"k8s-lx1036/k8s-ui/dashboard/controllers/resource/common/dataselect"
	"net/http"
)

type NamespaceController struct {

}

func (controller *NamespaceController) ListNamespaces() gin.HandlerFunc {
	return func(context *gin.Context) {
		k8sClient := client.DefaultClientManager.Client()
		dataSelectQuery := dataselect.ParseDataSelectFromRequest(context)
		namespaceList, _ := ListNamespacesByQuery(k8sClient, dataSelectQuery)
		
		context.JSON(http.StatusOK, common.JsonResponse{
			Errno: 0,
			Errmsg: "success",
			Data: namespaceList,
		})
	}
}
