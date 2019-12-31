package routers

import (
	"github.com/astaxie/beego"
	controllers2 "k8s-lx1036/app/framework/beego/controllers"
)

func init() {
	beego.Router("/", &controllers2.MainController{})
}
