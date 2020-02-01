package main

import (
	"k8s-lx1036/k8s-ui/backend/initial"
	routers_gin "k8s-lx1036/k8s-ui/backend/routers-gin"
	_ "k8s-lx1036/k8s-ui/backend/database/lorm"
)

const Version = "1.6.1"

func main() {
	/*cmd.Version = Version
	_ = cmd.RootCmd.Execute()*/

	//cmd2.Run()

	//database.InitDb()

	// K8S Client
	//initial.InitClient()

	// 初始化 rsa key
	initial.InitRsaKey()

	router := routers_gin.SetupRouter()

	_ = router.Run(":8080")
}
