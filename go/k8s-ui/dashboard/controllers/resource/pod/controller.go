package pod

import (
	"github.com/gin-gonic/gin"
	"k8s-lx1036/k8s-ui/dashboard/client"
)

type PodController struct {
}

func (controller *PodController) List() gin.HandlerFunc {
	return func(context *gin.Context) {
		k8sClient := client.DefaultClientManager.Client()
	}
}
