package main

import (
	"flag"
	"k8s-lx1036/k8s/network/loadbalancer/bgplb/pkg/controller/service"

	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
)

func main() {
	var (
		//port       = flag.Int("port", 7472, "HTTP listening port for Prometheus metrics")
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file (only needed when running outside of k8s)")
	)
	flag.Parse()

	restConfig, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		klog.Fatal(err)
	}

	service.New(restConfig)

}
