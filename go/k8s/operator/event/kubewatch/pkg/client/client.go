package client

import (
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	KubeClient kubernetes.Interface
)

func GetKubeClient(kubeconfig string) kubernetes.Interface {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err.Error(),
		}).Error("[kubeconfig]")
		return nil
	}

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err.Error(),
		}).Error("[kubeclient]")
		return nil
	}

	return client
}
