package main

import (
	"flag"
	"k8s.io/component-base/logs"

	"k8s-lx1036/k8s/network/cilium/cilium-ipam/pkg/controller/nodeipam"

	genericapiserver "k8s.io/apiserver/pkg/server"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
)

var (
	kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file (only needed when running outside of k8s)")
	ipam       = flag.String("ipam", "kubernetes", "cilium ipam mode, including 'kubernetes' or 'crd' mode")
)

// go run . --kubeconfig=`echo $HOME`/.kube/config
func main() {
	logs.InitLogs()
	defer logs.FlushLogs()

	flag.Parse()

	restConfig, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		klog.Fatal(err)
	}

	c := nodeipam.New(restConfig, *ipam)
	c.Run(genericapiserver.SetupSignalContext(), 1)
}
