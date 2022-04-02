package main

import (
	"flag"
	"k8s-lx1036/k8s/network/loadbalancer/bgplb/pkg/controller/speaker"
	"os"
	"runtime"

	"k8s-lx1036/k8s/network/loadbalancer/bgplb/pkg/apis/bgplb.k9s.io/v1"

	genericapiserver "k8s.io/apiserver/pkg/server"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/component-base/logs"
	"k8s.io/klog/v2"
)

func init() {
	_ = v1.AddToScheme(scheme.Scheme)
}

var (
	kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file (only needed when running outside of k8s)")
	grpcHosts  = flag.String("grpcHosts", ":50051", "specify the hosts that gobgpd listens on")
)

func main() {
	logs.InitLogs()
	defer logs.FlushLogs()

	if len(os.Getenv("GOMAXPROCS")) == 0 {
		runtime.GOMAXPROCS(runtime.NumCPU())
	}

	flag.Parse()

	restConfig, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		klog.Fatal(err)
	}

	c := speaker.NewSpeakerController(restConfig, *grpcHosts)
}
