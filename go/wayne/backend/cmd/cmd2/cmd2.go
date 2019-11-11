package cmd2

import "github.com/astaxie/beego"

func Run()  {
	if beego.BConfig.RunMode == "dev" {

	}

	beego.Run()
}
