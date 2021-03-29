package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"k8s-lx1036/k8s/scheduler/demo/extend-scheduler/pkg/plugins/priorityclass"

	"k8s.io/component-base/logs"
	"k8s.io/kubernetes/cmd/kube-scheduler/app"
)

// debug in local: `make dev`
// debug in idea ide: Program arguments加上 --config=/Users/liuxiang/Code/lx1036/code/go/k8s/scheduler/demo/extend-scheduler/scheduler-config.yaml --v=3
func main() {
	rand.Seed(time.Now().UTC().UnixNano())

	command := app.NewSchedulerCommand(
		app.WithPlugin(priorityclass.Name, priorityclass.New),
	)

	logs.InitLogs()
	defer logs.FlushLogs()

	if err := command.Execute(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
