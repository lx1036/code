package node

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedv1 "k8s.io/client-go/kubernetes/typed/core/v1"
	listerv1 "k8s.io/client-go/listers/core/v1"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
)

type nodeKey string

type Controller struct {
	kubeClient   kubernetes.Interface
	events       record.EventRecorder
	queue        workqueue.RateLimitingInterface
	nodeInformer cache.SharedIndexInformer
	nodeLister   listerv1.NodeLister
	syncFuncs    []cache.InformerSynced
}

func New(restConfig *restclient.Config) *Controller {
	kubeClient := kubernetes.NewForConfigOrDie(restConfig)

	broadcaster := record.NewBroadcaster()
	broadcaster.StartStructuredLogging(0)
	broadcaster.StartRecordingToSink(&typedv1.EventSinkImpl{Interface: kubeClient.CoreV1().Events("")})
	recorder := broadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: "cilium-node-ipam-controller"})
	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())

	c := &Controller{
		kubeClient: kubeClient,
		events:     recorder,
		queue:      queue,
	}

	factory := informers.NewSharedInformerFactory(kubeClient, 0)
	c.nodeInformer = factory.Core().V1().Nodes().Informer()
	c.nodeLister = factory.Core().V1().Nodes().Lister()
	c.nodeInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err == nil {
				c.queue.Add(nodeKey(key))
			}
		},
		DeleteFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err == nil {
				c.queue.Add(nodeKey(key))
			}
		},
	})

	c.syncFuncs = append(c.syncFuncs, c.nodeInformer.HasSynced, c.ippoolInformer.HasSynced)

}

func (c *Controller) Run(ctx context.Context, workers int) {
	defer runtime.HandleCrash()
	defer c.queue.ShutDown()

	klog.Info("Starting service controller")
	defer klog.Info("Shutting down service controller")

}
