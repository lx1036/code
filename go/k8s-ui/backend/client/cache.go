package client

import (
	"k8s-lx1036/k8s-ui/backend/client/api"
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
	stop := make(chan struct{})
	sharedInformerFactory := informers.NewSharedInformerFactory(client, defaultResyncPeriod)
	// Start all Resources defined in KindToResourceMap
	for _, value := range api.KindToResourceMap {
		genericInformer, err := sharedInformerFactory.ForResource(value.GroupVersionResourceKind.GroupVersionResource)
		if err != nil {
			return nil, err
		}
		go genericInformer.Informer().Run(stop)
	}

	sharedInformerFactory.Start(stop)

	return &CacheFactory{
		stopChan: stop,
		sharedInformerFactory:sharedInformerFactory,
	}, nil
}
