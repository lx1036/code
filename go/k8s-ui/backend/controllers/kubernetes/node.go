package kubernetes

import "github.com/gin-gonic/gin"

type KubeNodeController struct {
}

func (controller *KubeNodeController) PodStatistics() gin.HandlerFunc {
	return func(context *gin.Context) {

	}
}
