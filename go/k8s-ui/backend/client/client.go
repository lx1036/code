package client

import (
	"fmt"
	"github.com/astaxie/beego/logs"
	"k8s-lx1036/k8s-ui/backend/database/lorm"
	"k8s-lx1036/k8s-ui/backend/models"
	"k8s-lx1036/k8s-ui/backend/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"strconv"
	"sync"
	"time"
)

const (
	// High enough QPS to fit all expected use cases.
	defaultQPS = 1e6
	// High enough Burst to fit all expected use cases.
	defaultBurst = 1e6
	// full resyc cache resource time
	defaultResyncPeriod = 30 * time.Second
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
	var newClusters []models.Cluster
	err := lorm.DB.Where("status=?", models.ClusterStatusNormal).Find(&newClusters).Error
	if err != nil {
		logs.Error("empty clusters.")
		return
	}

	changed := clusterChanged(newClusters)
	if changed {
		logs.Info("cluster changed, so resync info...")
		// build new clientManager
		for i := 0; i < len(newClusters); i++ {
			cluster := newClusters[i]
			if cluster.Master == "" {
				logs.Warning("cluster's master is null:%s", cluster.Name)
				continue
			}
			clientSet, config, err := buildClient(cluster.Master, cluster.KubeConfig)
			if err != nil {
				logs.Warning("build cluster (%s) client error :%v", cluster.Name, err)
				continue
			}
			cacheFactory, err := buildCacheController(clientSet)
			if err != nil {
				logs.Warning("build cluster (%s) cache controller error :%v", cluster.Name, err)
				continue
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
	if util.SyncMapLen(clusterManagerSets) != len(clusters) {
		logs.Info("cluster length (%d) changed to (%d).", util.SyncMapLen(clusterManagerSets), len(clusters))
		return true
	}

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
	//configV1 := clientcmdapiv1.Config{}
	//err := json.Unmarshal([]byte(kubeconfig), &configV1)
	//if err != nil {
	//	logs.Error("json unmarshal kubeconfig error: %v ", err)
	//	return nil, nil, err
	//}

	//fmt.Printf("ClientCertificate: %s\n", configV1.AuthInfos[0].AuthInfo.ClientCertificate)

	//configObject, err := clientcmdlatest.Scheme.ConvertToVersion(&configV1, clientcmdapiv1.SchemeGroupVersion)
	//configInternal := configObject.(*clientcmdapi.Config)
	//clientConfig, err := clientcmd.NewDefaultClientConfig(*configInternal, &clientcmd.ConfigOverrides{
	//	AuthInfo:        clientcmdapi.AuthInfo{},
	//	ClusterDefaults: clientcmdapi.Cluster{Server: master},
	//	ClusterInfo:     clientcmdapi.Cluster{},
	//	Context:         clientcmdapi.Context{},
	//	CurrentContext:  "",
	//	Timeout:         "",
	//}).ClientConfig()
	clientConfig, err := clientcmd.RESTConfigFromKubeConfig([]byte(kubeconfig))
	if err != nil {
		logs.Error("build client config error. %v ", err)
		return nil, nil, err
	}

	clientConfig.QPS = defaultQPS
	clientConfig.Burst = defaultBurst
	clientSet, err := kubernetes.NewForConfig(clientConfig)
	if err != nil {
		logs.Error("(%s) kubernetes.NewForConfig(%v) error.%v", master, err, clientConfig)
		return nil, nil, err
	}

	pods, err := clientSet.CoreV1().Pods("kube-system").List(metav1.ListOptions{
		//LabelSelector: "deployment-abc",
	})
	if err != nil {
		panic(err.Error())
	}
	fmt.Println("numbers of pods: " + strconv.Itoa(len(pods.Items)))
	var containers []string
	for _, pod := range pods.Items {
		podName := pod.Name
		containers = append(containers, podName)
	}
	fmt.Printf("%s\n", containers)

	return clientSet, clientConfig, nil
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
