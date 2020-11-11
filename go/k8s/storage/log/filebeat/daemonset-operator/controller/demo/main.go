package main

import (
	"flag"
	"k8s-lx1036/k8s/storage/log/filebeat/daemonset-operator/controller/kubernetes"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/workqueue"
	"os"
	"sync"
	"time"
	log "github.com/sirupsen/logrus"
)

var (
	kubeconfig, apiServerURL string
	queue workqueue.RateLimitingInterface
)

func init() {
	flag.StringVar(&kubeconfig, "kubeconfig", "",
		"Paths to a kubeconfig. Only required if out-of-cluster.")
	
	// This flag is deprecated, it'll be removed in a future iteration, please switch to --kubeconfig.
	flag.StringVar(&apiServerURL, "master", "",
		"(Deprecated: switch to `--kubeconfig`) The address of the Kubernetes API server. Overrides any value in kubeconfig. "+
			"Only required if out-of-cluster.")
}

func main() {
	log.SetOutput(os.Stdout)
	log.SetFormatter(&log.JSONFormatter{})
	
	log.SetLevel(log.DebugLevel)
	
	flag.Parse()
	
	
	testQueue()
}

func testQueue()  {
	apiserverClient, err := kubernetes.NewApiServerClient(kubeconfig)
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}
	watcher, err := kubernetes.NewWatcher(apiserverClient, &kubernetes.Pod{}, kubernetes.WatchOptions{
		SyncTimeout: 10 * time.Minute,
		Node:        "master01",
		Namespace:   "default",
		IsUpdated:   nil,
	}, nil)
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}
	
	
	queue = workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "pods")
	
	
	watcher.AddEventHandler(kubernetes.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			pod := obj.(*kubernetes.Pod)
			log.Debug("Adding kubernetes pod: %s/%s", pod.GetNamespace(), pod.GetName())
			queue.Add(pod)
		},
		UpdateFunc: func(obj interface{}) {
			pod := obj.(*kubernetes.Pod)
			log.Debug("Updating kubernetes pod: %s/%s", pod.GetNamespace(), pod.GetName())
			queue.Add(pod)
		},
		DeleteFunc: func(obj interface{}) {
			pod := obj.(*kubernetes.Pod)
			log.Debug("Deleting kubernetes pod: %s/%s", pod.GetNamespace(), pod.GetName())
			queue.Add(pod)
		},
	})
	
	if err := watcher.Start(); err != nil {
		log.Error(err)
		os.Exit(1)
	}
	
	
	stop := make(chan struct{})
	go syncTask(queue, stop)
	
	
	<- stop
}

func syncTask(queue workqueue.RateLimitingInterface, stopCh chan struct{})  {
	wait.Until(func() {
		for process() {
		
		}
	}, 30 * time.Second, stopCh)
}

var lock = &sync.Mutex{}

func process() bool {
	lock.Lock()
	
	l := queue.Len()
	var objs []interface{}
	for i := 0; i < l; i++ {
		obj, quit := queue.Get()
		if quit {
			continue
		}
		
		objs = append(objs, obj)
	}
	
	batchSync(objs)
	
	for _, obj := range objs {
		queue.Done(obj)
	}
	
	lock.Unlock()
	
	return true
}

func batchSync(objs []interface{})  {
	for _, obj := range objs {
		pod := obj.(*kubernetes.Pod)
		log.Info(pod.Name)
	}
}
