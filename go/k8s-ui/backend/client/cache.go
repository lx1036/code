package client

import (
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	v1 "k8s.io/client-go/listers/core/v1"
)

type CacheFactory struct {
	stopChan              chan struct{}
	sharedInformerFactory informers.SharedInformerFactory
}

func (cache *CacheFactory)PodLister() v1.PodLister {
	return cache.sharedInformerFactory.Core().V1().Pods().Lister()
}

func (c ClusterManager) Close() {
	close(c.CacheFactory.stopChan)
}

func buildCacheController(client *kubernetes.Clientset) (*CacheFactory, error) {
	return &CacheFactory{}, nil
}
