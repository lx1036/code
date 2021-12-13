package main

import (
	goflag "flag"
	"github.com/spf13/pflag"
	"k8s-lx1036/k8s/kubelet/cri-hook-server/cmd/server/app"
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/component-base/logs"
	"math/rand"
	"os"
	"time"
)

// debug in local: go run . --config=./config.yaml
func main() {
	rand.Seed(time.Now().UnixNano())
	// init command context
	command := app.NewCriHookServerCommand()

	pflag.CommandLine.SetNormalizeFunc(cliflag.WordSepNormalizeFunc)
	pflag.CommandLine.AddGoFlagSet(goflag.CommandLine)
	logs.InitLogs()
	defer logs.FlushLogs()

	if err := command.Execute(); err != nil {
		os.Exit(1)
	}
}
