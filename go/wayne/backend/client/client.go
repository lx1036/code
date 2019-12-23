package client

import (
	"github.com/astaxie/beego/logs"
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


func BuildApiServerClient() {
	newClusters, err := models.ClusterModel.GetAllNormal()
	if err != nil {

	}

	changed := clusterChanged(newClusters)
	if changed {
		logs.Info("cluster changed, so resync info...")


		// build new clientManager
		for i := 0; i < len(newClusters); i++ {
			cluster := newClusters[i]

			clientSet, config, err := buildClient(cluster.Master, cluster.KubeConfig)
			if err != nil {

			}


			clusterManager := &ClusterManager{
				Client:       clientSet,
				Config:       config,
				Cluster:      &cluster,
				CacheFactory: cacheFactory,
				KubeClient:   NewResourceHandler(clientSet, cacheFactory),
			}

			managerInterface, ok := clusterManagerSets.Load(cluster.Name)
			if ok {
				manager := managerInterface.(*ClusterManager)
				manager.Close()
			}

			clusterManagerSets.Store(cluster.Name, clusterManager)

		}

		logs.Info("resync cluster finished! ")
	}
}

func clusterChanged(clusters []models.Cluster) bool {

	for _, cluster := range clusters {
		managerInterface, ok := clusterManagerSets.Load(cluster.Name)
		if !ok {
			// maybe add new cluster
			return true
		}
		manager := managerInterface.(*ClusterManager)
		// master changed, the cluster is changed, ignore others
		if manager.Cluster.Master != cluster.Master {
			logs.Info("cluster master (%s) changed to (%s).", manager.Cluster.Master, cluster.Master)
			return true
		}
		if manager.Cluster.Status != cluster.Status {
			logs.Info("cluster status (%d) changed to (%d).", manager.Cluster.Status, cluster.Status)
			return true
		}

		if manager.Cluster.KubeConfig != manager.Cluster.KubeConfig {
			logs.Info("cluster kubeConfig (%d) changed to (%d).", manager.Cluster.KubeConfig, cluster.KubeConfig)
			return true
		}
	}

	return false
}

func buildClient(master string, kubeconfig string) (*kubernetes.Clientset, *rest.Config, error) {

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

func Managers() *sync.Map {
	return clusterManagerSets
}


