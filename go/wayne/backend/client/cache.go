package client

import (
	"k8s.io/client-go/informers"
)

type CacheFactory struct {
	stopChan              chan struct{}
	sharedInformerFactory informers.SharedInformerFactory
}

func (c ClusterManager) Close() {
	close(c.CacheFactory.stopChan)
}
