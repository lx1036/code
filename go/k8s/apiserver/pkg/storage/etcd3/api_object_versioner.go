package etcd3

import (
	"k8s.io/apimachinery/pkg/runtime"
)

// APIObjectVersioner implements versioning and extracting etcd node information
// for objects that have an embedded ObjectMeta or ListMeta field.
type APIObjectVersioner struct{}

func (A APIObjectVersioner) UpdateObject(obj runtime.Object, resourceVersion uint64) error {
	panic("implement me")
}

func (A APIObjectVersioner) UpdateList(obj runtime.Object, resourceVersion uint64, continueValue string, remainingItemCount *int64) error {
	panic("implement me")
}

func (A APIObjectVersioner) PrepareObjectForStorage(obj runtime.Object) error {
	panic("implement me")
}

func (A APIObjectVersioner) ObjectResourceVersion(obj runtime.Object) (uint64, error) {
	panic("implement me")
}

func (A APIObjectVersioner) ParseResourceVersion(resourceVersion string) (uint64, error) {
	panic("implement me")
}
