package routers

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context/param"
)

func init() {
	const AppController = "k8s-lx1036/k8s-ui/backend/controllers/app:AppController"
	beego.GlobalControllerRouter[AppController] = append(
		beego.GlobalControllerRouter[AppController],
		beego.ControllerComments{
			Method:           "List",
			Router:           `/`,
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil,
		},
	)
}
