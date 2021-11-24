package main

import (
	"flag"
	"fmt"
	"k8s-lx1036/k8s/network/cilium/metallb/pkg/k8s/types"
	"os"
	"path/filepath"

	"k8s-lx1036/k8s/network/cilium/metallb/pkg/config"
	"k8s-lx1036/k8s/network/cilium/metallb/pkg/speaker"

	"k8s.io/klog/v2"
)

func main() {
	var (
		//port       = flag.Int("port", 7472, "HTTP listening port for Prometheus metrics")
		//name       = flag.String("name", "lb-ippool", "configmap name in default namespace")
		path = flag.String("config", "", "config file")
		//kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file (only needed when running outside of k8s)")
	)

	flag.Parse()
	if len(*path) == 0 {
		klog.Fatalf(fmt.Sprintf("config file is required"))
	}

	getSpeaker(*path)

}

func getSpeaker(path string) *speaker.Speaker {
	s, err := speaker.NewSpeaker(speaker.Config{
		MyNode: "",
		SList:  nil,
	})

	file, _ := filepath.Abs(path)
	f, err := os.Open(file)
	if err != nil {
		klog.Fatal(err)
	}
	c, err := config.Parse(f)
	if err != nil {
		klog.Fatal(err)
	}

	// 设置 ip pool
	if s.SetConfig(c) == types.SyncStateError {
		klog.Fatalf(fmt.Sprintf("failed to set config"))
	}

	return s
}
