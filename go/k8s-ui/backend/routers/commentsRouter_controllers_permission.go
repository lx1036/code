package routers

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context/param"
)

func init() {
	const NamespaceUserController = "k8s-lx1036/k8s-ui/backend/controllers/permission:NamespaceUserController"
	beego.GlobalControllerRouter[NamespaceUserController] = append(
		beego.GlobalControllerRouter[NamespaceUserController],
		beego.ControllerComments{
			Method:           "GetPermissionByNS",
			Router:           `/permissions/:id`,
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil,
		},
	)
}
