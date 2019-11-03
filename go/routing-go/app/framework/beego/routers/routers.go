package routers

import (
    "github.com/astaxie/beego"
    "k8s-lx1036/routing-go/app/framework/beego/controllers"
)

func init()  {
    beego.Router("/", &controllers.MainController{})
}
