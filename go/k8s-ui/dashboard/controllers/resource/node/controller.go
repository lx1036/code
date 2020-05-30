package node

import "github.com/gin-gonic/gin"

type NodeController struct {
}

func (controller *NodeController) ListNodes() gin.HandlerFunc {
	return func(context *gin.Context) {

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
