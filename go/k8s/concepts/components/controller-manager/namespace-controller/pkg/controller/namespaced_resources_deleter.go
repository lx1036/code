package controller

import (
	"fmt"

	"k8s-lx1036/k8s/concepts/components/controller-manager/namespace-controller/pkg/debug"

	log "github.com/sirupsen/logrus"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
)

type NamespacedResourcesDeleterInterface interface {
	Delete(namespace string) error
}

type namespacedResourcesDeleter struct {
	discoverResourcesFn func() ([]*metav1.APIResourceList, error)
}

func NewNamespacedResourcesDeleter(
	discoverResourcesFn func() ([]*metav1.APIResourceList, error),
) NamespacedResourcesDeleterInterface {

	deleter := &namespacedResourcesDeleter{
		discoverResourcesFn: discoverResourcesFn,
	}

	deleter.initOpCache()

	return deleter
}

func (deleter *namespacedResourcesDeleter) initOpCache() {
	resources, err := deleter.discoverResourcesFn()
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("unable to get all supported resources from server: %v", err))
	}
	if len(resources) == 0 {
		log.Fatalf("Unable to get any supported resources from server: %v", err)
	}

	debug.LogAPIResourceList(resources)

}

func (deleter *namespacedResourcesDeleter) Delete(namespace string) error {
	return nil
}
