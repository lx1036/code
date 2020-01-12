package routers

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context/param"
)

// namespace: https://kubernetes.io/docs/concepts/overview/working-with-objects/namespaces/

func init() {
	const controller = "k8s-lx1036/k8s-ui/backend/controllers/kubernetes"
	beego.GlobalControllerRouter[controller] = append(
		beego.GlobalControllerRouter[controller],
		beego.ControllerComments{
			Method:           "Create",
			Router:           `/:name/clusters/:cluster`,
			Filters:          nil,
			ImportComments:   nil,
			FilterComments:   nil,
			AllowHTTPMethods: []string{"post"},
			Params:           nil,
			MethodParams:     param.Make(),
		},
		beego.ControllerComments{
			Method:           "Resources",
			Router:           `/:namespaceid([0-9]+)/resources`,
			Filters:          nil,
			ImportComments:   nil,
			FilterComments:   nil,
			AllowHTTPMethods: []string{"get"},
			Params:           nil,
			MethodParams:     param.Make(),
		},
		beego.ControllerComments{
			Method:           "Statistics",
			Router:           `/:namespaceid([0-9]+)/statistics`,
			Filters:          nil,
			ImportComments:   nil,
			FilterComments:   nil,
			AllowHTTPMethods: []string{"get"},
			Params:           nil,
			MethodParams:     param.Make(),
		})
}
