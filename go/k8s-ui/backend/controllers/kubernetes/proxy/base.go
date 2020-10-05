package proxy

import (
	"k8s-lx1036/k8s-ui/backend/common/kubeclient"
)

type BaseController struct {
}

func (controller *BaseController) KubeClient(cluster string) (kubeclient.ResourceHandler, error) {
	clusterManager, err := kubeclient.Manager(cluster)
	if err != nil {
		return nil, err
	}

	return clusterManager.ResourceHandler, nil
}
