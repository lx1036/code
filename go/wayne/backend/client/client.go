package client

import (
	"k8s-lx1036/wayne/backend/models"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sync"
)

type ClusterManager struct {
	Cluster *models.Cluster
	// Deprecated: use KubeClient instead
	Client *kubernetes.Clientset
	// Deprecated: use KubeClient instead
	CacheFactory *CacheFactory
	Config       *rest.Config
	KubeClient   ResourceHandler
}


func BuildApiserverClient() {

}

func Client(cluster string) (*kubernetes.Clientset, error) {
	manager, err := Manager(cluster)
	if err != nil {
		return nil, err
	}
	return manager.Client, nil
}

var (
	clusterManagerSets = &sync.Map{}
)
func Manager(cluster string) (*ClusterManager, error) {
	managerInterface, exist := clusterManagerSets.Load(cluster)
	if !exist {
	
	}
	
	manager := managerInterface.(*ClusterManager)
	
	return manager, nil
}
