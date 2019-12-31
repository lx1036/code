package client

import (
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
)

type CacheFactory struct {
	stopChan              chan struct{}
	sharedInformerFactory informers.SharedInformerFactory
}

func (c ClusterManager) Close() {
	close(c.CacheFactory.stopChan)
}


func buildCacheController(client *kubernetes.Clientset) (*CacheFactory, error) {

}
