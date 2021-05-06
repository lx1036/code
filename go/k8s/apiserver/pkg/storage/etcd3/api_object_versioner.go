package etcd3

import (
	"strconv"

	"k8s.io/apimachinery/pkg/api/meta"
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

// PrepareObjectForStorage clears resource version and self link prior to writing to etcd.
func (A APIObjectVersioner) PrepareObjectForStorage(obj runtime.Object) error {
	accessor, err := meta.Accessor(obj)
	if err != nil {
		return err
	}
	accessor.SetResourceVersion("")
	accessor.SetSelfLink("")
	return nil
}

func (A APIObjectVersioner) ObjectResourceVersion(obj runtime.Object) (uint64, error) {
	accessor, err := meta.Accessor(obj)
	if err != nil {
		return 0, err
	}
	version := accessor.GetResourceVersion()
	if len(version) == 0 {
		return 0, nil
	}

	return strconv.ParseUint(version, 10, 64)
}

func (A APIObjectVersioner) ParseResourceVersion(resourceVersion string) (uint64, error) {
	panic("implement me")
}
