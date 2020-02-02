package routers

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context/param"
)

func init() {
	const BaseConfigController = "k8s-lx1036/k8s-ui/backend/controllers:BaseConfigController"
	beego.GlobalControllerRouter[BaseConfigController] = append(
		beego.GlobalControllerRouter[BaseConfigController],
		beego.ControllerComments{
			Method:           "ListBase",
			Router:           `/`,
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil,
		},
	)
}
