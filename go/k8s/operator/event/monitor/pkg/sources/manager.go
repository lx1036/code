package sources

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"go.opencensus.io/resource"
	"k8s-lx1036/k8s/plugins/event/monitor/pkg/sources/client"
	"k8s-lx1036/k8s/plugins/event/monitor/pkg/sources/events"
	"k8s-lx1036/k8s/plugins/event/monitor/pkg/sources/resources"
	"k8s-lx1036/k8s/plugins/event/monitor/pkg/utils"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"strconv"
	"time"
)

const (
	// Number of object pointers. Big enough so it won't be hit anytime soon with reasonable GetNewEvents frequency.
	EventsBufferSize = 100000
)

var (
	EventsBuffer = make(chan *events.Event, EventsBufferSize)
)

type Manager struct {
	//EventSource EventSource

	KubeClient       kubernetes.Interface
	WatchedResources []resources.Kind
}

type EventSource interface {
	GetNewEvents() *events.EventBatch
}

type manager struct {
	KubeClient kubernetes.Interface
}

func (manager *Manager) GetNewEvents() *events.EventBatch {
	result := events.EventBatch{
		Events:    nil,
		Timestamp: time.Now(),
	}

eventLoop:
	for {
		select {
		case event := <-EventsBuffer:
			result.Events = append(result.Events, event)
		default:
			break eventLoop
		}
	}

	return &result
}

func (manager *Manager) Start(stopChan chan struct{}) error {
	sharedInformerFactory := informers.NewSharedInformerFactory(manager.KubeClient, time.Minute*2)
	for kind, resource := range resources.Resources {
		if !resources.Find(manager.WatchedResources, kind) {
			continue
		}

		informer, err := sharedInformerFactory.ForResource(resource)
		if err != nil {
			return err
		}

		r := NewResource(informer)

		go informer.Informer().Run(stopChan)
		if !cache.WaitForCacheSync(stopChan, informer.Informer().HasSynced) {

		}

		wait.Until(r.runWorker, time.Second, stopChan)
	}

	stopCh := make(chan struct{})
	sharedInformerFactory.Start(stopCh)
	sharedInformerFactory.WaitForCacheSync(stopChan)
}

type Resource struct {
	Queue    workqueue.RateLimitingInterface
	Informer informers.GenericInformer
}

func NewResource(informer informers.GenericInformer) *Resource {
	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
	informer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			name, err := cache.MetaNamespaceKeyFunc(obj)
			event := events.Event{
				Name:         name,
				EventType:    events.Create,
				Namespace:    "",
				ResourceType: resource.Resource,
			}
			if err == nil {
				queue.Add(&event)
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {

		},
		DeleteFunc: func(obj interface{}) {

		},
	})

	return &Resource{
		Queue:    queue,
		Informer: informer,
	}
}
func (r *Resource) runWorker() {
	for r.processNextItem() {

	}
}
func (r *Resource) processNextItem() bool {
	event, quit := r.Queue.Get()
	if quit {
		return false
	}

	defer r.Queue.Done(event)

	err := r.processItem(event.(*events.Event))
	if err == nil {

	}

	return true
}

func (r *Resource) processItem(event *events.Event) error {
	obj, _, err := r.Informer.Informer().GetIndexer().GetByKey(event.Name)
	if err != nil {
		return fmt.Errorf("error fetching object with key %s from store: %v", event.Name, err)
	}

	// 针对resourceType分别过滤event
	objectMeta := utils.GetObjectMetaData(obj)
	switch event.EventType {
	case events.Create:
		EventsBuffer <- event
	case events.Update:

	case events.Delete:

	}

	fmt.Println(objectMeta)
	return nil
}

func NewManager(kubeconfig string) (*Manager, error) {
	kubeClient, err := client.GetKubeClient(kubeconfig)
	if err != nil {
		return nil, err
	}

	watchedResources, err := resources.GetWatchedResources()
	if err != nil {
		return nil, err
	}
	return &Manager{
		KubeClient:       kubeClient,
		WatchedResources: watchedResources,
	}, nil
}
