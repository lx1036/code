package controllers

import "github.com/astaxie/beego"

type MainController struct {
	beego.Controller
}

func (this *MainController) Get() {
	/*this.Data["Website"] = "lx1036.com"
	  this.Data["Email"] = "lx1036@126.com"
	  this.TplName = "index.tpl"*/

	this.Ctx.WriteString("hello world")
}
