package routers

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context/param"
)

func init() {
	const KubeNamespaceController = "k8s-lx1036/k8s-ui/backend/controllers/kubernetes/namespace:KubeNamespaceController"
	beego.GlobalControllerRouter[KubeNamespaceController] = append(
		beego.GlobalControllerRouter[KubeNamespaceController],
		beego.ControllerComments{
			Method:           "Resources",
			Router:           `/:namespaceId([0-9]+)/resources`,
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil,
		},
	)
}
