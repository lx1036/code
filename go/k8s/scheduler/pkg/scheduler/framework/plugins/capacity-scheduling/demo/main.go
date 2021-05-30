package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	_ "k8s-lx1036/k8s/scheduler/pkg/scheduler/apis/scheduling/scheme"
	capacity_scheduling "k8s-lx1036/k8s/scheduler/pkg/scheduler/framework/plugins/capacity-scheduling"

	"k8s.io/component-base/logs"
	"k8s.io/kubernetes/cmd/kube-scheduler/app"
)

// debug in local: `make dev`
// debug in idea ide: Program arguments加上 --config=/Users/liuxiang/Code/lx1036/code/go/k8s/scheduler/pkg/scheduler/framework/plugins/capacity-scheduling/demo/scheduler-config.yaml --v=3
func main() {
	rand.Seed(time.Now().UTC().UnixNano())
	
	command := app.NewSchedulerCommand(
		app.WithPlugin(capacity_scheduling.Name, capacity_scheduling.New),
	)

	logs.InitLogs()
	defer logs.FlushLogs()

	if err := command.Execute(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
