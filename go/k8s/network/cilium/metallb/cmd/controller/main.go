package main

import (
	"flag"
	"fmt"
	"k8s-lx1036/k8s/network/cilium/metallb/pkg/config"
	"os"
	"path/filepath"

	"k8s-lx1036/k8s/network/cilium/metallb/pkg/allocator"
	"k8s-lx1036/k8s/network/cilium/metallb/pkg/controller"
	"k8s-lx1036/k8s/network/cilium/metallb/pkg/k8s/types"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
)

// INFO: IPAM

type svcKey string
type cmKey string

// go run . --kubeconfig=`echo $HOME`/.kube/config --v=2 --config=`pwd`/config.yaml
func main() {
	var (
		//port       = flag.Int("port", 7472, "HTTP listening port for Prometheus metrics")
		name       = flag.String("name", "lb-ippool", "configmap name in default namespace")
		path       = flag.String("config", "", "config file")
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file (only needed when running outside of k8s)")
	)
	flag.Parse()
	if len(*path) == 0 {
		klog.Fatalf(fmt.Sprintf("config file is required"))
	}

	c := getIPAM(*path)

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
	_, cmInformer := cache.NewIndexerInformer(cmWatcher, &v1.ConfigMap{}, 0, cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err == nil {
				queue.Add(cmKey(key))
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(newObj)
			if err == nil {
				queue.Add(cmKey(key))
			}
		},
		DeleteFunc: func(obj interface{}) {
			key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			if err == nil {
				queue.Add(cmKey(key))
			}
		},
	}, cache.Indexers{})

	svcWatcher := cache.NewListWatchFromClient(clientset.CoreV1().RESTClient(), "services",
		metav1.NamespaceAll, fields.Everything())
	svcIndexer, svcInformer := cache.NewIndexerInformer(svcWatcher, &v1.Service{}, 0, cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err == nil {
				queue.Add(svcKey(key))
			}
		},
		UpdateFunc: func(old interface{}, new interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(new)
			if err == nil {
				queue.Add(svcKey(key))
			}
		},
		DeleteFunc: func(obj interface{}) {
			key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			if err == nil {
				queue.Add(svcKey(key))
			}
		},
	}, cache.Indexers{})
	epWatcher := cache.NewListWatchFromClient(clientset.CoreV1().RESTClient(), "endpoints",
		metav1.NamespaceAll, fields.Everything())
	epIndexer, epInformer := cache.NewIndexerInformer(epWatcher, &v1.Endpoints{}, 0, cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err == nil {
				queue.Add(svcKey(key))
			}
		},
		UpdateFunc: func(old interface{}, new interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(new)
			if err == nil {
				queue.Add(svcKey(key))
			}
		},
		DeleteFunc: func(obj interface{}) {
			key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			if err == nil {
				queue.Add(svcKey(key))
			}
		},
	}, cache.Indexers{})

	stopCh := make(chan struct{})
	go cmInformer.Run(stopCh)
	go svcInformer.Run(stopCh)
	if !cache.WaitForCacheSync(stopCh, cmInformer.HasSynced, svcInformer.HasSynced, epInformer.HasSynced) {
		klog.Fatalf(fmt.Sprintf("time out waiting for cache sync"))
	}

	sync := func(key interface{}, queue workqueue.RateLimitingInterface) types.SyncState {
		defer queue.Done(key)

		switch k := key.(type) {
		case svcKey:
			svc, exists, err := svcIndexer.GetByKey(string(k))
			if err != nil {
				klog.Errorf("failed to get service")
				return types.SyncStateError
			}
			if !exists {
				return c.SetBalancer(string(k), nil, nil)
			}

			if svc.(*v1.Service).Spec.Type != v1.ServiceTypeLoadBalancer {
				return types.SyncStateSuccess
			}

			endpoints, exists, err := epIndexer.GetByKey(string(k))
			if err != nil {
				klog.Errorf("failed to get endpoints")
				return types.SyncStateError
			}
			if !exists {
				return c.SetBalancer(string(k), nil, nil)
			}

			recorder.Eventf(svc.(*v1.Service), v1.EventTypeNormal, "SetBalancer", "update svc")
			return c.SetBalancer(string(k), svc.(*v1.Service), endpoints.(*v1.Endpoints))

		case cmKey:

			return types.SyncStateSuccess
		default:
			panic(fmt.Sprintf("unknown key type for %s %T", key, key))
		}
	}

	for {
		key, quit := queue.Get()
		if quit {
			return
		}

		state := sync(key, queue)
		switch state {
		case types.SyncStateSuccess:
			queue.Forget(key)
		case types.SyncStateError:
			queue.AddRateLimited(key)
		case types.SyncStateReprocessAll:
			queue.Forget(key)
		}
	}
}

func getIPAM(path string) *controller.Controller {
	ipam := &controller.Controller{
		IPs: allocator.New(),
	}

	file, _ := filepath.Abs(path)
	f, err := os.Open(file)
	if err != nil {
		klog.Fatal(err)
	}
	c, err := config.Parse(f)
	if err != nil {
		klog.Fatal(err)
	}

	ipam.SetConfig(c) // 设置 ip pool

	return ipam
}
