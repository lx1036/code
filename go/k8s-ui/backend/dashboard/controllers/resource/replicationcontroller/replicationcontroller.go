package replicationcontroller

import "github.com/gin-gonic/gin"

type ReplicationController struct {
}

func (controller *ReplicationController) List() gin.HandlerFunc {
	return func(context *gin.Context) {

	}
}
