package main

import (
	"flag"
	"fmt"

	"k8s-lx1036/k8s/network/cilium/metallb/pkg/allocator"
	"k8s-lx1036/k8s/network/cilium/metallb/pkg/controller"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
)

// SyncState is the result of calling synchronization callbacks.
type SyncState int

const (
	// The update was processed successfully.
	SyncStateSuccess SyncState = iota
	// The update caused a transient error, the k8s client should
	// retry later.
	SyncStateError
	// The update was accepted, but requires reprocessing all watched
	// services.
	SyncStateReprocessAll
)

type svcKey string
type cmKey string

func main() {
	var (
		port       = flag.Int("port", 7472, "HTTP listening port for Prometheus metrics")
		name       = flag.String("name", "name", "Kubernetes ConfigMap containing MetalLB's configuration")
		configNS   = flag.String("config-ns", "", "config file namespace (only needed when running outside of k8s)")
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file (only needed when running outside of k8s)")
	)
	flag.Parse()

	restConfig, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		klog.Fatal(err)
	}
	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		klog.Fatal(err)
	}

	broadcaster := record.NewBroadcaster()
	broadcaster.StartRecordingToSink(&corev1.EventSinkImpl{Interface: corev1.New(clientset.CoreV1().RESTClient()).Events("")})
	recorder := broadcaster.NewRecorder(scheme.Scheme, v1.EventSource{Component: "lb-controller"})
	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())

	cmWatcher := cache.NewListWatchFromClient(clientset.CoreV1().RESTClient(), "configmaps",
		metav1.NamespaceDefault, fields.OneTermEqualSelector("metadata.name", *name))
	cmIndexer, cmInformer := cache.NewIndexerInformer(cmWatcher, &v1.ConfigMap{}, 0, cache.ResourceEventHandlerFuncs{}, cache.Indexers{})

	svcWatcher := cache.NewListWatchFromClient(clientset.CoreV1().RESTClient(), "services",
		metav1.NamespaceAll, fields.Everything())
	svcIndexer, svcInformer := cache.NewIndexerInformer(svcWatcher, &v1.Service{}, 0, cache.ResourceEventHandlerFuncs{}, cache.Indexers{})

	stopCh := make(chan struct{})
	go cmInformer.Run(stopCh)
	go svcInformer.Run(stopCh)
	if !cache.WaitForCacheSync(stopCh, cmInformer.HasSynced, svcInformer.HasSynced) {
		klog.Fatalf(fmt.Sprintf("time out waiting for cache sync"))
	}

	c := &controller.Controller{
		IPs: allocator.New(),
	}

	sync := func(key interface{}, queue workqueue.RateLimitingInterface) SyncState {
		defer queue.Done(key)

		switch k := key.(type) {
		case svcKey:
			svc, exists, err := svcIndexer.GetByKey(string(k))
			if err != nil {
				klog.Errorf("failed to get service")
				return SyncStateError
			}
			if !exists {
				return c.SetBalancer(string(k), nil, nil)
			}

		}

	}

	for {
		key, quit := queue.Get()
		if quit {
			return
		}

		state := sync(key, queue)
		switch state {
		case SyncStateSuccess:
			queue.Forget(key)
		case SyncStateError:
			updateErrors.Inc()
			queue.AddRateLimited(key)
		case SyncStateReprocessAll:
			queue.Forget(key)
			ForceSync()
		}
	}
}
