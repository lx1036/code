package main

import (
	"os"

	"k8s-lx1036/k8s/scheduler/cmd/app"

	"k8s.io/component-base/cli"
	_ "k8s.io/component-base/logs/json/register" // for JSON log format registration
	_ "k8s.io/component-base/metrics/prometheus/clientgo"
	_ "k8s.io/component-base/metrics/prometheus/version" // for version metric registration
)

// debug in local: go run . --kubeconfig=`echo $HOME`/.kube/config --config=`pwd`/scheduler-config.yaml --v=3
func main() {
	command := app.NewSchedulerCommand()
	code := cli.Run(command)
	os.Exit(code)
}
