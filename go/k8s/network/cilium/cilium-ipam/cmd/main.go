package main

import (
	"flag"

	"k8s-lx1036/k8s/network/cilium/cilium-ipam/pkg/controller/nodeipam"

	genericapiserver "k8s.io/apiserver/pkg/server"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
)

// go run . --kubeconfig=`echo $HOME`/.kube/config
func main() {
	var (
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file (only needed when running outside of k8s)")
	)
	flag.Parse()

	restConfig, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		klog.Fatal(err)
	}

	c := nodeipam.New(restConfig)
	c.Run(genericapiserver.SetupSignalContext(), 1)
}
