package namespace

import (
	"k8s-lx1036/k8s-ui/backend/common/kubeclient"
	"k8s-lx1036/k8s-ui/backend/common/resources/common"
)

// ResourcesUsageByNamespace Count resource usage for a namespace
func ResourcesUsageByNamespace(cli kubeclient.ResourceHandler, namespace, selector string) (*common.ResourceList, error) {
	return &common.ResourceList{}, nil
}
