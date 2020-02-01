package cmd2

import (
	"github.com/astaxie/beego"
	"k8s-lx1036/k8s-ui/backend/initial"
	_ "k8s-lx1036/k8s-ui/backend/routers"
)

func Run() {
	if beego.BConfig.RunMode == "dev" {
		beego.BConfig.WebConfig.DirectoryIndex = true
		beego.BConfig.WebConfig.StaticDir["/swagger"] = "swagger"
	}

	initial.InitDb()

	// K8S Client
	initial.InitClient()

	// 初始化 rsa key
	initial.InitRsaKey()

	beego.Run()
}
