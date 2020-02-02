package cmd2

import (
	"github.com/astaxie/beego"
	database "k8s-lx1036/k8s-ui/backend/database/initial"
	"k8s-lx1036/k8s-ui/backend/initial"
	_ "k8s-lx1036/k8s-ui/backend/routers"
)

func Run() {
	if beego.BConfig.RunMode == "dev" {
		beego.BConfig.WebConfig.DirectoryIndex = true
		beego.BConfig.WebConfig.StaticDir["/swagger"] = "swagger"
	}

	database.InitDb()

	// K8S Client
	initial.InitClient()

	// 初始化 rsa key
	initial.InitRsaKey()

	beego.Run()
}
