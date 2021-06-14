package main

import (
	"flag"
	"os"
	"runtime"

	"k8s-lx1036/k8s/bigdata/spark-on-k8s/spark-operator/cmd/app"

	genericapiserver "k8s.io/apiserver/pkg/server"
	"k8s.io/component-base/logs"
)

// debug in local: go run . --kubeconfig=`echo $HOME`/.kube/config --v=2
func main() {
	logs.InitLogs()
	defer logs.FlushLogs()

	if len(os.Getenv("GOMAXPROCS")) == 0 {
		runtime.GOMAXPROCS(runtime.NumCPU())
	}

	cmd := app.NewSparkOperatorCommand(genericapiserver.SetupSignalHandler())
	cmd.Flags().AddGoFlagSet(flag.CommandLine)
	if err := cmd.Execute(); err != nil {
		panic(err)
	}
}
