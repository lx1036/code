package main

import (
	"k8s-lx1036/k8s/kubelet/pkg/cadvisor"

	"k8s.io/klog/v2"
)

// GOOS=linux GOARCH=amd64 go build .
func main() {
	client, err := cadvisor.New(nil, "/var/lib/kubelet", []string{"/kubepods"}, true)
	if err != nil {
		panic(err)
	}

	machineInfo, err := client.MachineInfo()
	if err != nil {
		panic(err)
	}

	klog.Infof("%d cores", machineInfo.NumCores)
}
