package cmd2

import (
	"fmt"
	"github.com/astaxie/beego"
	"k8s-lx1036/k8s-ui/backend/initial"
	_ "k8s-lx1036/k8s-ui/backend/routers"
)

func Run() {
	initial.InitDb()

	if beego.BConfig.RunMode == "dev" {

	}

	// K8S Client
	initial.InitClient()

	// 初始化 rsa key
	//initial.InitRsaKey()

	fmt.Println(beego.BConfig.RunMode)

	beego.Run()
}
