package kubernetes

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"time"
)

func Run(clientSet kubernetes.Interface, InformerResources []schema.GroupVersionResource, stopChannel chan struct{})  {
	sharedInformerFactory := informers.NewSharedInformerFactory(clientSet, time.Second*10)
	for _, resource := range InformerResources {
		// Informer: informer作为异步事件处理框架，完成了事件监听和分发处理两个过程
		genericInformer, err := sharedInformerFactory.ForResource(resource)
		if err != nil {
			panic(err)
		}
		go genericInformer.Informer().Run(stopChannel)
	}
	sharedInformerFactory.Start(stopChannel)
}
