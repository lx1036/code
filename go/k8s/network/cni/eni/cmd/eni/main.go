package main

import (
	"flag"
	"fmt"
	"os"

	"k8s-lx1036/k8s/network/cni/eni/pkg/daemon"

	"k8s.io/klog/v2"
)

const defaultConfigPath = "/etc/eni/eni.json"
const defaultPidPath = "/var/run/eni/eni.pid"
const defaultSocketPath = "/var/run/eni/eni.socket"
const debugSocketPath = "unix:///var/run/eni/eni_debug.socket"

var (
	daemonMode string
	kubeconfig string
)

func main() {
	flagSet := flag.NewFlagSet("eni", flag.ExitOnError)
	flagSet.StringVar(&daemonMode, "daemon-mode", "VPC", "eni network mode")
	flagSet.StringVar(&kubeconfig, "kubeconfig", "", "path to kubeconfig file")
	err := flagSet.Parse(os.Args[1:])
	if err != nil {
		panic(err)
	}

	if err := daemon.Run(defaultPidPath, defaultSocketPath, readonlyListen, defaultConfigPath, kubeconfig, master, daemonMode); err != nil {
		klog.Fatalf(fmt.Sprintf("[main]run daemon err: %v", err))
	}
}
