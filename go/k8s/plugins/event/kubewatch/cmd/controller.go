package cmd

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"k8s-lx1036/k8s/plugins/event/kubewatch/pkg/event"
	"k8s-lx1036/k8s/plugins/event/monitor/pkg/utils"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

// Event indicate the informerEvent
type Event struct {
	key          string
	eventType    string
	namespace    string
	resourceType string
}

// Controller object
type Controller struct {
	logger    *logrus.Entry
	clientset kubernetes.Interface
	queue     workqueue.RateLimitingInterface
	informer  cache.SharedIndexInformer
	//eventHandler handlers.Handler
}

var serverStartTime time.Time

const maxRetries = 5

func Start(config Config, kubeClient kubernetes.Interface) {
	// Adding Default Critical Alerts
	// For Capturing Critical Event NodeNotReady in Nodes
	nodeNotReadyInformer := cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
				options.FieldSelector = "involvedObject.kind=Node,type=Normal,reason=NodeNotReady"
				return kubeClient.CoreV1().Events(config.Namespace).List(context.TODO(), options)
			},
			WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
				options.FieldSelector = "involvedObject.kind=Node,type=Normal,reason=NodeNotReady"
				return kubeClient.CoreV1().Events(config.Namespace).Watch(context.TODO(), options)
			},
		},
		&apiv1.Event{},
		0, //Skip resync
		cache.Indexers{},
	)
	nodeNotReadyController := newResourceController(kubeClient, nodeNotReadyInformer, "NodeNotReady")
	stopNodeNotReadyCh := make(chan struct{})
	defer close(stopNodeNotReadyCh)
	go nodeNotReadyController.Run(stopNodeNotReadyCh)

	// For Capturing Critical Event NodeReady in Nodes
	nodeReadyInformer := cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
				options.FieldSelector = "involvedObject.kind=Node,type=Normal,reason=NodeReady"
				return kubeClient.CoreV1().Events(config.Namespace).List(context.TODO(), options)
			},
			WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
				options.FieldSelector = "involvedObject.kind=Node,type=Normal,reason=NodeReady"
				return kubeClient.CoreV1().Events(config.Namespace).Watch(context.TODO(), options)
			},
		},
		&apiv1.Event{},
		0, //Skip resync
		cache.Indexers{},
	)
	nodeReadyController := newResourceController(kubeClient, nodeReadyInformer, "NodeReady")
	stopNodeReadyCh := make(chan struct{})
	defer close(stopNodeReadyCh)
	go nodeReadyController.Run(stopNodeReadyCh)

	/*informer := cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
				return kubeClient.CoreV1().Pods(config.Namespace).List(context.TODO(),options)
			},
			WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
				return kubeClient.CoreV1().Pods(config.Namespace).Watch(context.TODO(),options)
			},
		},
		&apiv1.Pod{},
		0, //Skip resync
		cache.Indexers{},
	)
	c := newResourceController(kubeClient, informer, "pod")
	stopCh := make(chan struct{})
	defer close(stopCh)
	go c.Run(stopCh)*/

	// For Capturing CrashLoopBackOff Events in pods
	backoffInformer := cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
				options.FieldSelector = "involvedObject.kind=Pod,type=Warning,reason=BackOff"
				return kubeClient.CoreV1().Events(config.Namespace).List(context.TODO(), options)
			},
			WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
				options.FieldSelector = "involvedObject.kind=Pod,type=Warning,reason=BackOff"
				return kubeClient.CoreV1().Events(config.Namespace).Watch(context.TODO(), options)
			},
		},
		&apiv1.Event{},
		0, //Skip resync
		cache.Indexers{},
	)
	backoffcontroller := newResourceController(kubeClient, backoffInformer, "Backoff")
	stopBackoffCh := make(chan struct{})
	defer close(stopBackoffCh)
	go backoffcontroller.Run(stopBackoffCh)

	informer := cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
				return kubeClient.AppsV1().Deployments(config.Namespace).List(context.TODO(), options)
			},
			WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
				return kubeClient.AppsV1().Deployments(config.Namespace).Watch(context.TODO(), options)
			},
		},
		&appsv1.Deployment{},
		0, //Skip resync
		cache.Indexers{},
	)
	c := newResourceController(kubeClient, informer, "deployment")
	stopCh := make(chan struct{})
	defer close(stopCh)
	go c.Run(stopCh)

	nodeInformer := cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
				return kubeClient.CoreV1().Nodes().List(context.TODO(), options)
			},
			WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
				return kubeClient.CoreV1().Nodes().Watch(context.TODO(), options)
			},
		},
		&apiv1.Node{},
		0, //Skip resync
		cache.Indexers{},
	)
	nodeController := newResourceController(kubeClient, nodeInformer, "node")
	nodeStopCh := make(chan struct{})
	defer close(nodeStopCh)
	go nodeController.Run(nodeStopCh)

	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGTERM, syscall.SIGINT)
	<-sigterm
}

func newResourceController(client kubernetes.Interface, informer cache.SharedIndexInformer, resourceType string) *Controller {
	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
	var newEvent Event
	var err error
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			newEvent.key, err = cache.MetaNamespaceKeyFunc(obj)
			newEvent.eventType = "create"
			newEvent.resourceType = resourceType
			//logrus.WithField("pkg", "kubewatch-"+resourceType).Infof("Processing add to %v: %s", resourceType, newEvent.key)
			if err == nil {
				queue.Add(newEvent)
			}
		},
		UpdateFunc: func(old, new interface{}) {
			newEvent.key, err = cache.MetaNamespaceKeyFunc(old)
			newEvent.eventType = "update"
			newEvent.resourceType = resourceType
			//logrus.WithField("pkg", "kubewatch-"+resourceType).Infof("Processing update to %v: %s", resourceType, newEvent.key)
			if err == nil {
				queue.Add(newEvent)
			}
		},
		DeleteFunc: func(obj interface{}) {
			newEvent.key, err = cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			newEvent.eventType = "delete"
			newEvent.resourceType = resourceType
			newEvent.namespace = utils.GetObjectMetaData(obj).Namespace
			//logrus.WithField("pkg", "kubewatch-"+resourceType).Infof("Processing delete to %v: %s", resourceType, newEvent.key)
			if err == nil {
				queue.Add(newEvent)
			}
		},
	})

	return &Controller{
		logger:    logrus.WithField("pkg", "kubewatch-"+resourceType),
		clientset: client,
		informer:  informer,
		queue:     queue,
		//eventHandler: eventHandler,
	}
}

// Run starts the kubewatch controller
func (c *Controller) Run(stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()
	defer c.queue.ShutDown()

	c.logger.Info("Starting kubewatch controller")
	serverStartTime = time.Now().Local()

	go c.informer.Run(stopCh)

	if !cache.WaitForCacheSync(stopCh, c.HasSynced) {
		utilruntime.HandleError(fmt.Errorf("timed out waiting for caches to sync"))
		return
	}

	c.logger.Info("Kubewatch controller synced and ready")

	wait.Until(c.runWorker, time.Second, stopCh)
}

// HasSynced is required for the cache.Controller interface.
func (c *Controller) HasSynced() bool {
	return c.informer.HasSynced()
}
func (c *Controller) runWorker() {
	for c.processNextItem() {
		// continue looping
	}
}

func (c *Controller) processNextItem() bool {
	newEvent, quit := c.queue.Get()

	if quit {
		return false
	}
	defer c.queue.Done(newEvent)
	err := c.processItem(newEvent.(Event))
	if err == nil {
		// No error, reset the ratelimit counters
		c.queue.Forget(newEvent)
	} else if c.queue.NumRequeues(newEvent) < maxRetries {
		c.logger.Errorf("Error processing %s (will retry): %v", newEvent.(Event).key, err)
		c.queue.AddRateLimited(newEvent)
	} else {
		// err != nil and too many retries
		c.logger.Errorf("Error processing %s (giving up): %v", newEvent.(Event).key, err)
		c.queue.Forget(newEvent)
		utilruntime.HandleError(err)
	}

	return true
}

/* TODOs
- Enhance event creation using client-side cacheing machanisms - pending
- Enhance the processItem to classify events - done
- Send alerts correspoding to events - done
*/
func (c *Controller) processItem(newEvent Event) error {
	obj, _, err := c.informer.GetIndexer().GetByKey(newEvent.key)
	if err != nil {
		return fmt.Errorf("Error fetching object with key %s from store: %v", newEvent.key, err)
	}
	// get object's metedata
	objectMeta := utils.GetObjectMetaData(obj)
	if objectMeta.CreationTimestamp.Sub(serverStartTime).Seconds() < 0 {
		return nil
	}

	// hold status type for default critical alerts
	var status string

	// namespace retrived from event key incase namespace value is empty
	if newEvent.namespace == "" && strings.Contains(newEvent.key, "/") {
		substring := strings.Split(newEvent.key, "/")
		newEvent.namespace = substring[0]
		newEvent.key = substring[1]
	}

	// process events based on its type
	switch newEvent.eventType {
	case "create":
		// compare CreationTimestamp and serverStartTime and alert only on latest events
		// Could be Replaced by using Delta or DeltaFIFO
		if objectMeta.CreationTimestamp.Sub(serverStartTime).Seconds() > 0 {
			switch newEvent.resourceType {
			case "NodeNotReady":
				status = "Danger"
			case "NodeReady":
				status = "Normal"
			case "NodeRebooted":
				status = "Danger"
			case "Backoff":
				status = "Danger"
			default:
				status = "Normal"
			}
			kbEvent := event.Event{
				Name:      objectMeta.Name,
				Namespace: newEvent.namespace,
				Kind:      newEvent.resourceType,
				Status:    status,
				Reason:    "Created",
			}
			//c.eventHandler.Handle(kbEvent)

			logrus.WithFields(logrus.Fields{
				"Name":      kbEvent.Name,
				"Namespace": kbEvent.Namespace,
				"Kind":      kbEvent.Kind,
				"Status":    kbEvent.Status,
				"Reason":    kbEvent.Reason,
			}).Info("[Created]")

			return nil
		}
	case "update":
		/* TODOs
		- enahace update event processing in such a way that, it send alerts about what got changed.
		*/
		switch newEvent.resourceType {
		case "Backoff":
			status = "Danger"
		default:
			status = "Warning"
		}
		kbEvent := event.Event{
			Name:      newEvent.key,
			Namespace: newEvent.namespace,
			Kind:      newEvent.resourceType,
			Status:    status,
			Reason:    "Updated",
		}
		logrus.WithFields(logrus.Fields{
			"Name":      kbEvent.Name,
			"Namespace": kbEvent.Namespace,
			"Kind":      kbEvent.Kind,
			"Status":    kbEvent.Status,
			"Reason":    kbEvent.Reason,
		}).Info("[Updated]")

		//c.eventHandler.Handle(kbEvent)
		return nil
	case "delete":
		kbEvent := event.Event{
			Name:      newEvent.key,
			Namespace: newEvent.namespace,
			Kind:      newEvent.resourceType,
			Status:    "Danger",
			Reason:    "Deleted",
		}
		logrus.WithFields(logrus.Fields{
			"Name":      kbEvent.Name,
			"Namespace": kbEvent.Namespace,
			"Kind":      kbEvent.Kind,
			"Status":    kbEvent.Status,
			"Reason":    kbEvent.Reason,
		}).Info("[Deleted]")

		//c.eventHandler.Handle(kbEvent)
		return nil
	}
	return nil
}
