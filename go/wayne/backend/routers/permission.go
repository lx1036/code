package routers

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context/param"
)

/**
这些 comment routes 应该是 beego 自动生成的，但是在 GOPATH 外没法自动生成，
@see https://github.com/astaxie/beego/issues/3162 还得等官方修复
所以，选择手动添加。
 */
func init() {
	const AppUserController = "k8s-lx1036/wayne/backend/controllers/permission:AppUserController"
	beego.GlobalControllerRouter[AppUserController] = append(
		beego.GlobalControllerRouter[AppUserController],
		/**
		GET localhost:8080/api/v1/apps/12/users/3
		 */
		beego.ControllerComments{
			Method: "Get",
			Router: `/:id`,
			AllowHTTPMethods: []string{"get"},
			MethodParams: param.Make(),
			Filters: nil,
			Params: nil,
		},
	)
}
