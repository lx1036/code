package main

import (
	"k8s-lx1036/k8s-ui/backend/cmd/cmd2"
)

const Version = "1.6.1"

func main() {
	/*cmd.Version = Version
	_ = cmd.RootCmd.Execute()*/

	cmd2.Run()
}