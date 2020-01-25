package config

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"k8s-lx1036/k8s-ui/backend/common"
	"k8s-lx1036/k8s-ui/backend/controllers/base"
	"k8s-lx1036/k8s-ui/backend/models"
	"k8s-lx1036/k8s-ui/backend/util"
	"net/http"
	"strconv"
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
	configMap["appLabelKey"] = util.AppLabelKey
	configMap["namespaceLabelKey"] = util.NamespaceLabelKey
	configMap["enableRobin"] = beego.AppConfig.DefaultBool("EnableRobin", false)
	configMap["ldapLogin"] = parseAuthEnabled("auth.ldap")
	configMap["oauth2Login"] = parseAuthEnabled("auth.oauth2")
	configMap["enableApiKeys"] = beego.AppConfig.DefaultBool("EnableApiKeys", false)

	var configs []models.Config
	err := models.GetAll(new(models.Config), &configs, &common.QueryParam{
		PageNo:   1,
		PageSize: 1000,
	})
	if err != nil {
		logs.Error("list base configs error: %v", err)
		controller.Ctx.Output.SetStatus(http.StatusInternalServerError)
		return
	}

	for _, config := range configs {
		configMap[string(config.Name)] = config.Value
	}

	controller.Ctx.Output.SetStatus(http.StatusOK)
	controller.Data["json"] = base.Result{Data: configMap}
	controller.ServeJSON()
}

func parseAuthEnabled(name string) bool {
	enabledSection, err := beego.AppConfig.GetSection(name)
	if err != nil {
		return false
	}

	enabled, err := strconv.ParseBool(enabledSection["enabled"])
	if err != nil {
		return false
	}

	return enabled
}
