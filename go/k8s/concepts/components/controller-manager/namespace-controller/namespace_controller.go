package main

import (
	"os"

	"k8s-lx1036/k8s/concepts/components/controller-manager/namespace-controller/pkg/cmd"
)

// 测试NamespaceController:
// https://github.com/kubernetes/kubernetes/blob/release-1.17/pkg/controller/namespace/namespace_controller.go

// go run . --debug=true
func main() {
	if err := cmd.NewNamespaceControllerCommand().Execute(); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}
