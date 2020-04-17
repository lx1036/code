package kubernetes

import (
	"k8s-lx1036/k8s-ui/backend/kubernetes/plugins/event/k8s-event-monitor/common"
	kubeapi "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubewatch "k8s.io/apimachinery/pkg/watch"
	kubeclient "k8s.io/client-go/kubernetes"
	kubev1core "k8s.io/client-go/kubernetes/typed/core/v1"
	kuberest "k8s.io/client-go/rest"
	"net/url"
	"time"
)

const (
	EventBufferSize = 100000
)

type EventSource struct {
	EventClient kubev1core.EventInterface
	StopChannel chan struct{}
	// Export *kubeapi.Event to outside as buffer
	EventBuffer chan *kubeapi.Event
}

func (eventSource *EventSource) Watch() {
	for {
		eventList, err := eventSource.EventClient.List(metav1.ListOptions{})
		if err != nil {
			continue
		}

		watcher, err := eventSource.EventClient.Watch(metav1.ListOptions{
			ResourceVersion: eventList.ResourceVersion,
		})
		if err != nil {
			continue
		}

		watchChannel := watcher.ResultChan()

	watchLoop:
		for {
			select {
			case eventObj, ok := <-watchChannel:
				if !ok {

					break watchLoop
				}
				if eventObj.Type == kubewatch.Error {

					break watchLoop
				}
				if event, ok := eventObj.Object.(*kubeapi.Event); ok {
					switch eventObj.Type {
					case kubewatch.Added, kubewatch.Modified:
						select {
						case eventSource.EventBuffer <- event:
							// buffer not full
						default:
							// buffer is full, drop the event
						}
					case kubewatch.Deleted:

					case kubewatch.Bookmark:

					default:

					}
				}
			case <-eventSource.StopChannel:
				return
			}
		}

	}
}

func (eventSource *EventSource) GetEvents() common.Events {
	var events common.Events

readEventLoop:
	for {
		select {
		case event := <-eventSource.EventBuffer:
			events.Events = append(events.Events, event)
		default:
			break readEventLoop
		}
	}

	events.Timestamp = time.Now()

	return events
}

func (eventSource *EventSource) ListEvents()  {

}

func NewKubernetesEventSource(uri *url.URL) (common.EventSource, error) {
	kubeConfig, err := GetKubeClientConfig(uri)
	if err != nil {
		return nil, err
	}
	clientSet, err := kubeclient.NewForConfig(kubeConfig)
	if err != nil {
		return nil, err
	}

	// https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#event-v1-core
	eventClient := clientSet.CoreV1().Events(kubeapi.NamespaceDefault)
	eventSource := &EventSource{
		EventClient: eventClient,
		StopChannel: make(chan struct{}),
		EventBuffer: make(chan *kubeapi.Event, EventBufferSize),
	}

	go eventSource.Watch()

	return eventSource, nil
}

const (
	defaultInClusterConfig = true
)

func GetKubeClientConfig(uri *url.URL) (*kuberest.Config, error) {
	var kubeConfig *kuberest.Config
	var err error
	inClusterConfig := defaultInClusterConfig
	if inClusterConfig {
		kubeConfig, err = kuberest.InClusterConfig()
		if err != nil {
			return nil, err
		}
	}

	return kubeConfig, nil
}
