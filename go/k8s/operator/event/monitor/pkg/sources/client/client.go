package client

import (
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func GetKubeClient(kubeconfig string) (kubernetes.Interface, error) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err.Error(),
		}).Error("[kubeconfig]")
		return nil, err
	}

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err.Error(),
		}).Error("[kubeclient]")
		return nil, err
	}

	return client, nil
}
