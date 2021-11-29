package main

import (
	"flag"
	"fmt"
	"k8s-lx1036/k8s/network/cilium/metallb/pkg/k8s/types"
	"os"
	"path/filepath"

	"k8s-lx1036/k8s/network/cilium/metallb/pkg/config"
	"k8s-lx1036/k8s/network/cilium/metallb/pkg/speaker"

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

type svcKey string

// go run . --kubeconfig=`echo $HOME`/.kube/config --config=`pwd`/config.yaml
func main() {
	var (
		//port       = flag.Int("port", 7472, "HTTP listening port for Prometheus metrics")
		//name       = flag.String("name", "lb-ippool", "configmap name in default namespace")
		path       = flag.String("config", "", "config file")
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file (only needed when running outside of k8s)")
	)

	flag.Parse()
	if len(*path) == 0 {
		klog.Fatalf(fmt.Sprintf("config file is required"))
	}

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

	// INFO: (1) 与 router server 建立 bgp session
	s := getSpeaker(*path)

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
			//key, err := cache.MetaNamespaceKeyFunc(new)
			//if err == nil {
			//	//queue.Add(svcKey(key))
			//	klog.Infof(fmt.Sprintf("update %s", key))
			//}
		},
		DeleteFunc: func(obj interface{}) {
			//key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			//if err == nil {
			//	//queue.Add(svcKey(key))
			//	klog.Infof(fmt.Sprintf("delete %s", key))
			//}
		},
	}, cache.Indexers{})

	epWatcher := cache.NewListWatchFromClient(clientset.CoreV1().RESTClient(), "endpoints",
		metav1.NamespaceAll, fields.Everything())
	epIndexer, epInformer := cache.NewIndexerInformer(epWatcher, &v1.Endpoints{}, 0, cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			//key, err := cache.MetaNamespaceKeyFunc(obj)
			//if err == nil {
			//	//queue.Add(svcKey(key))
			//	klog.Info(key)
			//}
		},
		UpdateFunc: func(old interface{}, new interface{}) {
			//key, err := cache.MetaNamespaceKeyFunc(new)
			//if err == nil {
			//	klog.Info(key)
			//	//queue.Add(svcKey(key))
			//}
		},
		DeleteFunc: func(obj interface{}) {
			//key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			//if err == nil {
			//	//queue.Add(svcKey(key))
			//	klog.Info(key)
			//}
		},
	}, cache.Indexers{})

	stopCh := make(chan struct{})
	go svcInformer.Run(stopCh)
	go epInformer.Run(stopCh)
	if !cache.WaitForCacheSync(stopCh, svcInformer.HasSynced, epInformer.HasSynced) {
		klog.Fatalf(fmt.Sprintf("time out waiting for cache sync"))
	}

	sync := func(key interface{}, queue workqueue.RateLimitingInterface) error {
		defer queue.Done(key)

		switch k := key.(type) {
		case svcKey:
			svc, exists, err := svcIndexer.GetByKey(string(k))
			if err != nil {
				return err
			}
			if !exists {
				return fmt.Errorf("not exist")
			}
			endpoints, exists, err := epIndexer.GetByKey(string(k))
			if err != nil {
				klog.Errorf("failed to get endpoints")
				return err
			}
			if !exists {
				return fmt.Errorf("not exist")
			}

			if svc.(*v1.Service).Spec.Type != v1.ServiceTypeLoadBalancer {
				return nil
			}

			recorder.Eventf(svc.(*v1.Service), v1.EventTypeNormal, "SetBalancer", "advertise svc ip")
			s.SetBalancer(string(k), svc.(*v1.Service), endpoints.(*v1.Endpoints))
			return nil
		default:
			panic(fmt.Sprintf("unknown key type for %s %T", key, key))
		}
	}

	for {
		key, quit := queue.Get()
		if quit {
			return
		}

		err := sync(key, queue)
		if err != nil {
			klog.Error(err)
		} else {
			queue.Forget(key)
		}
	}
}

func getSpeaker(path string) *speaker.Speaker {
	s, err := speaker.NewSpeaker(speaker.Config{
		MyNode: "",
		SList:  nil,
	})

	file, _ := filepath.Abs(path)
	f, err := os.Open(file)
	if err != nil {
		klog.Fatal(err)
	}
	c, err := config.Parse(f)
	if err != nil {
		klog.Fatal(err)
	}

	// INFO: (1) 与 router server 建立 bgp session
	if s.SetConfig(c) == types.SyncStateError {
		klog.Fatalf(fmt.Sprintf("failed to set config"))
	}

	return s
}
