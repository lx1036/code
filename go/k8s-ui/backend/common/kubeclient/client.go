package kubeclient

import (
	"context"
	log "github.com/sirupsen/logrus"
	"k8s-lx1036/k8s-ui/backend/database"
	"k8s-lx1036/k8s-ui/backend/models"
	"k8s-lx1036/k8s-ui/backend/common/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	corev1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
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

var (
	clusterManagerSets = &sync.Map{}
)

type ClusterManager struct {
	Cluster         *models.Cluster
	Config          *rest.Config
	ResourceHandler ResourceHandler
}

func clusterChanged(clusters []models.Cluster) bool {
	if util.SyncMapLen(clusterManagerSets) != len(clusters) {
		log.Info("cluster length (%d) changed to (%d).", util.SyncMapLen(clusterManagerSets), len(clusters))
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
			log.Info("cluster master (%s) changed to (%s).", manager.Cluster.Master, cluster.Master)
			return true
		}
		if manager.Cluster.Status != cluster.Status {
			log.Info("cluster status (%d) changed to (%d).", manager.Cluster.Status, cluster.Status)
			return true
		}

		if manager.Cluster.KubeConfig != manager.Cluster.KubeConfig {
			log.Info("cluster kubeConfig (%d) changed to (%d).", manager.Cluster.KubeConfig, cluster.KubeConfig)
			return true
		}
	}

	return false
}

func buildClient(master string, kubeconfig string) (*kubernetes.Clientset, *rest.Config, error) {
	clientConfig, err := clientcmd.RESTConfigFromKubeConfig([]byte(kubeconfig))
	if err != nil {
		log.Error("build client config error. %v ", err)
		return nil, nil, err
	}

	clientConfig.QPS = defaultQPS
	clientConfig.Burst = defaultBurst
	clientSet, err := kubernetes.NewForConfig(clientConfig)
	if err != nil {
		log.Error("(%s) kubernetes.NewForConfig(%v) error.%v", master, err, clientConfig)
		return nil, nil, err
	}

	return clientSet, clientConfig, nil
}

func Client(cluster string) (*kubernetes.Clientset, error) {
	manager, err := Manager(cluster)
	if err != nil {
		return nil, err
	}
	return manager.ResourceHandler.GetClient(), nil
}

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

type CacheFactory struct {
	stopChan              chan struct{}
	sharedInformerFactory informers.SharedInformerFactory
}

func (cache *CacheFactory) PodLister() corev1.PodLister {
	return cache.sharedInformerFactory.Core().V1().Pods().Lister()
}

func (c ClusterManager) Close() {
	close(c.ResourceHandler.GetCacheFactory().stopChan)
}

func buildCacheController(client *kubernetes.Clientset) (*CacheFactory, error) {
	stop := make(chan struct{})
	sharedInformerFactory := informers.NewSharedInformerFactory(client, defaultResyncPeriod)
	// Start all Resources defined in KindToResourceMap
	for _, value := range KindToResourceMap {
		genericInformer, err := sharedInformerFactory.ForResource(value.GroupVersionResourceKind.GroupVersionResource)
		if err != nil {
			return nil, err
		}
		go genericInformer.Informer().Run(stop)
	}

	sharedInformerFactory.Start(stop)

	return &CacheFactory{
		stopChan:              stop,
		sharedInformerFactory: sharedInformerFactory,
	}, nil
}

type ResourceHandler interface {
	Create(kind string, namespace string, object *runtime.Unknown) (*runtime.Unknown, error)
	Update(kind string, namespace string, name string, object *runtime.Unknown) (*runtime.Unknown, error)
	Get(kind string, namespace string, name string) (runtime.Object, error)
	List(kind string, namespace string, labelSelector string) ([]runtime.Object, error)
	Delete(kind string, namespace string, name string, options *metav1.DeleteOptions) error
	GetClient() *kubernetes.Clientset
	GetCacheFactory() *CacheFactory
}

type resourceHandler struct {
	Client       *kubernetes.Clientset
	CacheFactory *CacheFactory
}

func (handler *resourceHandler) GetClient() *kubernetes.Clientset {
	return handler.Client
}

func (handler *resourceHandler) GetCacheFactory() *CacheFactory {
	return handler.CacheFactory
}

func (handler *resourceHandler) Create(kind string, namespace string, object *runtime.Unknown) (*runtime.Unknown, error) {
	var result *runtime.Unknown
	err := handler.Client.CoreV1().RESTClient().Post().Resource(kind).Namespace(namespace).Body(object).Do(context.TODO()).Into(result)
	
	return result, err
}

func (handler *resourceHandler) Update(kind string, namespace string, name string, object *runtime.Unknown) (*runtime.Unknown, error) {
	var result *runtime.Unknown
	err := handler.Client.CoreV1().RESTClient().Put().Resource(kind).Namespace(namespace).Name(name).Body(object).Do(context.TODO()).Into(result)
	
	return result, err
}

func (handler *resourceHandler) List(kind string, namespace string, labelSelector string) ([]runtime.Object, error) {
	panic("implement me")
}

func (handler *resourceHandler) Delete(kind string, namespace string, name string, options *metav1.DeleteOptions) error {
	return handler.Client.CoreV1().RESTClient().Delete().Namespace(namespace).Resource(kind).Name(name).Body(options).Do(context.TODO()).Error()
}

func (handler *resourceHandler) Get(kind string, namespace string, name string) (runtime.Object, error) {
	resource, ok := KindToResourceMap[kind]
	if !ok {

	}

	genericInformer, err := handler.CacheFactory.sharedInformerFactory.ForResource(resource.GroupVersionResourceKind.GroupVersionResource)
	if err != nil {
		return nil, err
	}

	var result runtime.Object
	lister := genericInformer.Lister()
	if resource.Namespaced {
		result, err = lister.ByNamespace(namespace).Get(name)
		if err != nil {
			return nil, err
		}
	} else {
		result, err = lister.Get(name)
		if err != nil {
			return nil, err
		}
	}

	result.GetObjectKind().SetGroupVersionKind(schema.GroupVersionKind{
		Group:   resource.GroupVersionResourceKind.Group,
		Version: resource.GroupVersionResourceKind.Version,
		Kind:    resource.GroupVersionResourceKind.Kind,
	})

	return result, nil
}

func NewResourceHandler(kubeClient *kubernetes.Clientset, cacheFactory *CacheFactory) ResourceHandler {
	return &resourceHandler{
		Client:       kubeClient,
		CacheFactory: cacheFactory,
	}
}

func BuildApiServerClient() {
	var newClusters []models.Cluster
	err := database.DB.Where("status=?", models.ClusterStatusNormal).Where("deleted_at is null").Find(&newClusters).Error
	if err != nil {
		log.Error("empty clusters.")
		return
	}

	changed := clusterChanged(newClusters)
	if changed {
		log.Info("cluster changed, so resync info...")
		// build new clientManager
		for i := 0; i < len(newClusters); i++ {
			cluster := newClusters[i]
			if cluster.Master == "" {
				log.Warning("cluster's master is null:%s", cluster.Name)
				continue
			}
			clientSet, config, err := buildClient(cluster.Master, cluster.KubeConfig)
			if err != nil {
				log.Warning("build cluster (%s) client error :%v", cluster.Name, err)
				continue
			}
			cacheFactory, err := buildCacheController(clientSet)
			if err != nil {
				log.Warning("build cluster (%s) cache controller error :%v", cluster.Name, err)
				continue
			}

			clusterManager := &ClusterManager{
				Config:          config,
				Cluster:         &cluster,
				ResourceHandler: NewResourceHandler(clientSet, cacheFactory),
			}
			managerInterface, ok := clusterManagerSets.Load(cluster.Name)
			if ok {
				manager := managerInterface.(*ClusterManager)
				manager.Close()
			}

			clusterManagerSets.Store(cluster.Name, clusterManager)
		}

		log.Info("resync cluster finished! ")
	}
}
