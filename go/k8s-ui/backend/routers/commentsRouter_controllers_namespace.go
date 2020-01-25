package routers

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context/param"
)

func init() {
	const NamespaceController = "k8s-lx1036/k8s-ui/backend/controllers/namespace:NamespaceController"
	beego.GlobalControllerRouter[NamespaceController] = append(
		beego.GlobalControllerRouter[NamespaceController],
		beego.ControllerComments{
			Method:           "Statistics",
			Router:           "/:namespaceId([0-9]+)/statistics",
			Filters:          nil,
			ImportComments:   nil,
			FilterComments:   nil,
			AllowHTTPMethods: []string{"get"},
			Params:           nil,
			MethodParams:     param.Make(),
		},
	)
}
