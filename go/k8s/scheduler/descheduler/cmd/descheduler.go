package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"k8s-lx1036/k8s/scheduler/descheduler/cmd/app"

	"k8s.io/component-base/logs"
)

// debug in local: `make dev`
func main() {
	rand.Seed(time.Now().UTC().UnixNano())

	command := app.NewDeschedulerCommand()

	logs.InitLogs()
	defer logs.FlushLogs()

	if err := command.Execute(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
