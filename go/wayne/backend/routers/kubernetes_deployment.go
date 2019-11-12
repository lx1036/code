package routers

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context/param"
)

func init() {
	const DeploymentController = "k8s-lx1036/wayne/backend/controllers/deployment:DeploymentController"
	beego.GlobalControllerRouter[DeploymentController] = append(
		beego.GlobalControllerRouter[DeploymentController],
		beego.ControllerComments{
			Method: "List",
			Router: `/`,
			AllowHTTPMethods: []string{"get"},
			MethodParams: param.Make(),
			Filters: nil,
			Params: nil,
		},
	)
}
