package proxy

import (
	"github.com/gin-gonic/gin"
	"k8s-lx1036/k8s-ui/backend/controllers/base"
	"net/http"
)

type KubeProxyController struct {
	BaseController
}

func (controller *KubeProxyController) Get() gin.HandlerFunc {
	return func(context *gin.Context) {
		//appId := context.Param("appId")
		cluster := context.Param("cluster")
		namespace := context.Param("namespace")
		kind := context.Param("kind")
		kindName := context.Param("kindName")

		resourceHandler, err := controller.KubeClient(cluster)
		if err != nil || resourceHandler == nil {

			return
		}

		result, err := resourceHandler.Get(kind, namespace, kindName)
		if err != nil {

		}

		context.JSON(http.StatusOK, base.JsonResponse{
			Errno:  0,
			Errmsg: "success",
			Data:   result,
		})
	}
}
