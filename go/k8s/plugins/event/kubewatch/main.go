package main

import (
	"k8s-lx1036/k8s/plugins/event/kubewatch/cmd"
)

// go run . --configfile=./monitor.yaml --kubeconfig=/Users/liuxiang/.kube/config
func main() {
	cmd.Execute()
}
