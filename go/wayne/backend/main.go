package main

import (
	"k8s-lx1036/wayne/backend/cmd/cmd2"
	_ "k8s-lx1036/wayne/backend/routers"
)

const Version = "1.6.1"

func main() {
	/*cmd.Version = Version
	_ = cmd.RootCmd.Execute()*/

	cmd2.Run()
}
