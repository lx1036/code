package main

import (
	"flag"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	kubeV1 "k8s.io/api/apps/v1"
	coreV1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"path/filepath"
	"time"
)

const (
	DefaultRestartCount = 3
)

var (
	kubeconfig  *string
	namespace   *string
	clientSet   *kubernetes.Clientset
	stopChannel = make(chan struct{})

	sharedInformerFactory informers.SharedInformerFactory
)

var InformerResources = []schema.GroupVersionResource{
	{
		Group:    kubeV1.GroupName,
		Version:  kubeV1.SchemeGroupVersion.Version,
		Resource: "deployments",
	},
	{
		Group:    coreV1.GroupName,
		Version:  coreV1.SchemeGroupVersion.Version,
		Resource: "pods",
	},
}

func SetupRouter() *gin.Engine {
	router := gin.Default()

	// curl localhost:8080
	router.GET("/", func(context *gin.Context) {
		deploymentGenericInformer, err := sharedInformerFactory.ForResource(schema.GroupVersionResource{
			Group:    kubeV1.GroupName,
			Version:  kubeV1.SchemeGroupVersion.Version,
			Resource: "deployments",
		})
		if err != nil {
			panic(err)
		}
		// sharedIndexInformer: 该成员indexer是一个保存全量数据的缓存Store
		deployments, err := deploymentGenericInformer.Lister().ByNamespace(*namespace).List(labels.Everything())
		for _, obj := range deployments {
			deployment, ok := obj.(*kubeV1.Deployment)
			if !ok {
				log.Warnf("failed to convert obj[%v] to pod", obj)
				continue
			}
			log.Infof("deployment name is %s", deployment.Name)
		}

		genericInformer, err := sharedInformerFactory.ForResource(schema.GroupVersionResource{
			Group:    coreV1.GroupName,
			Version:  coreV1.SchemeGroupVersion.Version,
			Resource: "pods",
		})
		if err != nil {
			panic(err)
		}
		var podRestartCount = 0
		pods, err := genericInformer.Lister().ByNamespace(*namespace).List(labels.Everything())
		for _, obj := range pods {
			pod, ok := obj.(*coreV1.Pod)
			if !ok {
				log.Warnf("failed to convert obj[%v] to pod", obj)
				continue
			}

			for _, containerStatuses := range pod.Status.ContainerStatuses {
				podRestartCount += int(containerStatuses.RestartCount)
			}
			log.Infof("pod[%s] restart count: %d", pod.Name, podRestartCount)
			if podRestartCount >= DefaultRestartCount {
				log.Warnf("pod[%s] restart count: %d", pod.Name, podRestartCount)
			}
		}
	})

	return router
}

func main() {
	log.SetOutput(os.Stdout)
	log.SetFormatter(&log.JSONFormatter{})

	namespace = flag.String("namespace", coreV1.NamespaceDefault, "scoped specific namespace")

	if home, _ := os.UserHomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "absolute path to kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to kubeconfig file")
	}

	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err)
	}
	clientSet, err = kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	sharedInformerFactory = informers.NewSharedInformerFactory(clientSet, time.Second*10)
	for _, resource := range InformerResources {
		// Informer: informer作为异步事件处理框架，完成了事件监听和分发处理两个过程
		genericInformer, err := sharedInformerFactory.ForResource(resource)
		if err != nil {
			panic(err)
		}
		go genericInformer.Informer().Run(stopChannel)
	}
	sharedInformerFactory.Start(stopChannel)

	router := SetupRouter()

	err = router.Run(":8080")
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Info("[app level]")
	}
}
