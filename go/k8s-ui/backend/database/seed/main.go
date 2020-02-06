package main

import (
	"k8s-lx1036/k8s-ui/backend/database/seed/cmd"
	"os"
)

// https://github.com/spf13/cobra/blob/master/cobra/main.go
func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
