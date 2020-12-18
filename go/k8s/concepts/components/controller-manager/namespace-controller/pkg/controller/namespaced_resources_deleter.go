package controller

import (
	"fmt"
	"sync"

	"k8s-lx1036/k8s/concepts/components/controller-manager/namespace-controller/pkg/debug"

	log "github.com/sirupsen/logrus"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
)

// operation is used for caching if an operation is supported on a dynamic client.
type operation string

const (
	operationDeleteCollection operation = "deletecollection"
	operationList             operation = "list"
	// assume a default estimate for finalizers to complete when found on items pending deletion.
	finalizerEstimateSeconds int64 = int64(15)
)

// operationKey is an entry in a cache.
type operationKey struct {
	operation operation
	gvr       schema.GroupVersionResource
}

type operationNotSupportedCache struct {
	lock sync.RWMutex
	m    map[operationKey]bool
}

func (o *operationNotSupportedCache) setNotSupported(key operationKey) {
	o.lock.Lock()
	defer o.lock.Unlock()
	o.m[key] = true
}

type NamespacedResourcesDeleterInterface interface {
	Delete(namespaceName string) error
}

type namespacedResourcesDeleter struct {
	discoverResourcesFn func() ([]*metav1.APIResourceList, error)

	opCache *operationNotSupportedCache
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

	// 过滤出verb只有list和deletecollection的resource
	var deletableGroupVersionResources []schema.GroupVersionResource
	for _, resource := range resources {
		gv, err := schema.ParseGroupVersion(resource.GroupVersion)
		if err != nil {
			log.Errorf("Failed to parse GroupVersion %q, skipping: %v", resource.GroupVersion, err)
			continue
		}

		for _, apiResource := range resource.APIResources {
			gvr := schema.GroupVersionResource{Group: gv.Group, Version: gv.Version, Resource: apiResource.Name}
			verbs := sets.NewString([]string(apiResource.Verbs)...)

			if !verbs.Has("delete") {
				log.Infof("Skipping resource %v because it cannot be deleted.", gvr)
			}

			// verb没有list和deletecollection的resource，cache到opCache对象
			for _, op := range []operation{operationList, operationDeleteCollection} {
				if !verbs.Has(string(op)) {
					deleter.opCache.setNotSupported(operationKey{operation: op, gvr: gvr})
				}
			}

			deletableGroupVersionResources = append(deletableGroupVersionResources, gvr)
		}
	}
}

// 删除给定namespace中的所有资源对象
// 删除前：检查namespace.DeletionTimestamp字段值；检查namespace.Status.Phase是不是"Terminating"状态
// 删除后：移除namespace.Spec.Finalizers并finalize下namespace
func (deleter *namespacedResourcesDeleter) Delete(namespaceName string) error {
	return nil
}

type ResourcesRemainingError struct {
	Estimate int64
}

func (e *ResourcesRemainingError) Error() string {
	return fmt.Sprintf("some content remains in the namespace, estimate %d seconds before it is removed", e.Estimate)
}
