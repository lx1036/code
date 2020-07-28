package csrf

import (
	"github.com/gin-gonic/gin"
	"golang.org/x/net/xsrftoken"
	"k8s-lx1036/k8s-ui/dashboard/controllers/resource/common/client/api"
	"net/http"
)

type CsrfController struct {
}

type CsrfToken struct {
	Token string `json:"token"`
}

func (controller *CsrfController) GetCsrfToken() gin.HandlerFunc {
	return func(context *gin.Context) {
		action := context.Param("action")
		token := xsrftoken.Generate(api.GenerateCsrfKey(), "none", action)
		context.JSON(http.StatusOK, gin.H{
			"errno":  0,
			"errmsg": "success",
			"data":   CsrfToken{Token: token},
		})
	}
}
