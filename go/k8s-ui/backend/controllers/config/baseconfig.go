package config

import (
	"github.com/astaxie/beego"
	"k8s-lx1036/k8s-ui/backend/controllers/base"
	"net/http"
)

type BaseConfigController struct {
	beego.Controller
}

func (controller *BaseConfigController) URLMapping() {
	controller.Mapping("ListBase", controller.ListBase)
}

// @router / [get]
func (controller *BaseConfigController) ListBase() {
	configMap := make(map[string]interface{})
	configMap["appUrl"] = beego.AppConfig.String("AppUrl")
	configMap["betaUrl"] = beego.AppConfig.String("BetaUrl")
	configMap["enableDBLogin"] = beego.AppConfig.DefaultBool("EnableDBLogin", false)

	controller.Ctx.Output.SetStatus(http.StatusOK)
	controller.Data["json"] = base.Result{Data: configMap}
	controller.ServeJSON()
}
