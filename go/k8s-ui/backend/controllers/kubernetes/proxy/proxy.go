package proxy

import "github.com/gin-gonic/gin"

type KubeProxyController struct {

}

func (controller *KubeProxyController) Get() gin.HandlerFunc {
	return func(context *gin.Context) {
		appId := context.Param("appId")
		cluster := context.Param("cluster")
		namespace := context.Param("namespace")
		kind := context.Param("kind")
		kindName := context.Param("kindName")
		
		
		
	}
}


