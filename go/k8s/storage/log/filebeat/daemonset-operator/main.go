package main

import (
	"flag"
	"k8s-lx1036/k8s/storage/log/filebeat/daemonset-operator/controller"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	log "github.com/sirupsen/logrus"
)

var (
	kubeconfig, apiServerURL string
)

func init() {
	flag.StringVar(&kubeconfig, "kubeconfig", "",
		"Paths to a kubeconfig. Only required if out-of-cluster.")
	
	// This flag is deprecated, it'll be removed in a future iteration, please switch to --kubeconfig.
	flag.StringVar(&apiServerURL, "master", "",
		"(Deprecated: switch to `--kubeconfig`) The address of the Kubernetes API server. Overrides any value in kubeconfig. "+
			"Only required if out-of-cluster.")
}

func main() {
	log.SetOutput(os.Stdout)
	log.SetFormatter(&log.JSONFormatter{})
	
	flag.Parse()
	
	kubeClient, err := CreateApiServerClient(kubeconfig)
	if err != nil {
		log.Fatalf("Failed to create kubernetes apiserver client: %v", err)
		os.Exit(1)
	}
	
	ctl := controller.New(kubeClient)
	
	ctl.Run()
}

func CreateApiServerClient(kubeconfig string) (*kubernetes.Clientset, error) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, err
	}
	
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	
	return client, nil
}
