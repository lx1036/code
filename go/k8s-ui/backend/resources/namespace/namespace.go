package namespace

import (
	"k8s-lx1036/k8s-ui/backend/client"
	"k8s-lx1036/k8s-ui/backend/resources/common"
)

// ResourcesUsageByNamespace Count resource usage for a namespace
func ResourcesUsageByNamespace(cli client.ResourceHandler, namespace, selector string) (*common.ResourceList, error) {

}
