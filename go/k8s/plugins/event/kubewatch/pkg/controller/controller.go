package controller

import (
	"k8s-lx1036/k8s/plugins/event/kubewatch/config"
	"k8s-lx1036/k8s/plugins/event/kubewatch/pkg/client"
	"k8s.io/client-go/tools/cache"
)

func Run(config *config.Config)  {
	kubeClient := client.GetKubeClient("")
	informer := cache.NewSharedIndexInformer()
	go informer.Run()
}


