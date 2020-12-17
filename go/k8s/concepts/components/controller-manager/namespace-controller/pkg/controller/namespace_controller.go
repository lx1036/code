package controller

import (
	"k8s-lx1036/k8s/concepts/components/controller-manager/namespace-controller/pkg/kube"
	"k8s.io/client-go/util/workqueue"
)

type NamespaceController struct {
	queue                      workqueue.RateLimitingInterface
	namespacedResourcesDeleter NamespacedResourcesDeleterInterface
}

func NewNamespaceController() *NamespaceController {
	clientset := kube.GetClientset()

	discoverResourcesFn := clientset.Discovery().ServerPreferredNamespacedResources

	namespaceController := &NamespaceController{
		queue:                      workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "namespace"),
		namespacedResourcesDeleter: NewNamespacedResourcesDeleter(discoverResourcesFn),
	}

	return namespaceController
}

func (controller *NamespaceController) Run(workers int, stopCh <-chan struct{}) error {

	return nil
}
