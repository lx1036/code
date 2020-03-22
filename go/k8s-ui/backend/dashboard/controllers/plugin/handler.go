package plugin

import (
	"github.com/gin-gonic/gin"
	"k8s-lx1036/k8s-ui/backend/dashboard/client"
)

type PluginController struct {
}

func (controller *PluginController) HandleConfig() gin.HandlerFunc {
	return func(context *gin.Context) {
		result := GetPluginList(client.DefaultClientManager.GetPluginClient(), "")
	}
}
