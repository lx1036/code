package main

import (
	"flag"
	log "github.com/sirupsen/logrus"
	"k8s-lx1036/k8s/storage/log/filebeat/daemonset-operator/common"
	"k8s-lx1036/k8s/storage/log/filebeat/daemonset-operator/controller"
	"os"
)

var (
	kubeconfig = flag.String("kubeconfig", "", "Paths to a kubeconfig. Only required if out-of-cluster.")
	host = flag.String("host", "", "Specified node")
	namespace = flag.String("namespace", "", "Specified namespace")
)



func main() {
	flag.Parse()
	
	log.SetOutput(os.Stdout)
	log.SetFormatter(&log.JSONFormatter{})

	options := &common.Options{
		KubeConfig: *kubeconfig,
		Host: *host,
		Namespace: *namespace,
	}

	ctl := controller.New(options)

	ctl.Run()
}

