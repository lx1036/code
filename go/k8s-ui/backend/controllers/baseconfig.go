package controllers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"k8s-lx1036/k8s-ui/backend/controllers/base"
	"k8s-lx1036/k8s-ui/backend/models"
	"net/http"
)

type BaseConfigController struct {
}

//func (controller *BaseConfigController) URLMapping() {
//	controller.Mapping("ListBase", controller.ListBase)
//}

//func (controller *BaseConfigController) Init() {
//
//}

func (controller *BaseConfigController) ListBase() gin.HandlerFunc {
	return func(context *gin.Context) {
		configMap := make(map[string]interface{})
		configMap["appUrl"] = viper.GetString("default.AppUrl")
		configMap["betaUrl"] = viper.GetString("default.BetaUrl")
		configMap["enableDBLogin"] = viper.GetBool("default.EnableDBLogin")
		configMap["appLabelKey"] = util.AppLabelKey
		configMap["namespaceLabelKey"] = util.NamespaceLabelKey
		configMap["enableRobin"] = viper.GetBool("default.EnableRobin")
		configMap["ldapLogin"] = viper.GetBool("auth.ldap.enabled")
		configMap["oauth2Login"] = viper.GetBool("auth.oauth2.enabled")
		configMap["enableApiKeys"] = viper.GetBool("default.EnableApiKeys")

		var configs []models.Config
		err := lorm.DB.Limit(100).Find(&configs).Error
		if err != nil {
			context.JSON(http.StatusInternalServerError, base.JsonResponse{
				Errno:  -1,
				Errmsg: fmt.Sprintf("failed: list base configs error: %s", err.Error()),
				Data:   nil,
			})
			return
		}

		for _, config := range configs {
			configMap[string(config.Name)] = config.Value
		}

		context.JSON(http.StatusOK, base.JsonResponse{
			Errno:  0,
			Errmsg: "success",
			Data:   configMap,
		})
	}
}
