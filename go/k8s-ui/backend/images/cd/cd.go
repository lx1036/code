package main

import (
	"k8s-lx1036/k8s-ui/backend/images/cd/cmd"
	"os"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
