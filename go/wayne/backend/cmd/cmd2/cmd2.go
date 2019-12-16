package cmd2

import (
	"fmt"
	"github.com/astaxie/beego"
	"k8s-lx1036/wayne/backend/initial"
)

func Run()  {
	initial.InitDb()


	if beego.BConfig.RunMode == "dev" {

	}

	// 初始化 rsa key
	initial.InitRsaKey()


	fmt.Println(beego.BConfig.RunMode)

	beego.Run()
}
