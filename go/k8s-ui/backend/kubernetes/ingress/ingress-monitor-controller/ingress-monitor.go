package main

import (
	"flag"
	log "github.com/sirupsen/logrus"
	kubeV1 "k8s.io/api/core/v1"
	kubeV1Beta1 "k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/workqueue"
	"os"
	"path/filepath"
	"time"
)

// ingress monitor
func main() {
	var kubeconfig *string
	if home, _ := os.UserHomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "absolute path to kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to kubeconfig file")
	}
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err)
	}
	kubeClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}
	restClient := kubeClient.ExtensionsV1beta1().RESTClient()
	resource := "ingresses"
	namespace := kubeV1.NamespaceAll

	resyncPeriod := time.Duration(3) * time.Second

	watcher := cache.NewListWatchFromClient(restClient, resource, namespace, fields.Everything())
	_, informer := cache.NewIndexerInformer(watcher, &kubeV1Beta1.Ingress{}, resyncPeriod, cache.ResourceEventHandlerFuncs{
		AddFunc:    onResourceAdd,
		UpdateFunc: onResourceUpdate,
		DeleteFunc: onResourceDelete,
	}, cache.Indexers{})

	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
	defer queue.ShutDown()

	stopChan := make(chan struct{})
	defer close(stopChan)

	go informer.Run(stopChan)

	if !cache.WaitForCacheSync(stopChan, informer.HasSynced) {
		os.Exit(1)
	}

	runWorker := func() {
		action, quit := queue.Get()
		if quit {
			return
		}
		defer queue.Done(action)
	}

	go wait.Until(runWorker, time.Second, stopChan)

	<-stopChan
	log.Info("Stopping Ingress Monitor Controller...")
}

func onResourceAdd(obj interface{}) {

}

func onResourceUpdate(oldObj, newObj interface{}) {

}

func onResourceDelete(obj interface{}) {

}
