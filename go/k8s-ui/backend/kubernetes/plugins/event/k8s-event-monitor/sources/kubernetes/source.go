package kubernetes

import (
	kubeapi "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubewatch "k8s.io/apimachinery/pkg/watch"
	kubeclient "k8s.io/client-go/kubernetes"
	kubev1core "k8s.io/client-go/kubernetes/typed/core/v1"
	"net/url"
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

func (eventSource *EventSource) GetEvents() []*kubeapi.Event {
	var events []*kubeapi.Event

readEventLoop:
	for {
		select {
		case event := <-eventSource.EventBuffer:
			events = append(events, event)
		default:
			break readEventLoop
		}
	}

	return events
}

func NewKubernetesEventSource(uri *url.URL) (*EventSource, error) {
	kubeConfig, err := GetKubeClientConfig(uri)
	if err != nil {
		return nil, err
	}
	clientSet, err := kubeclient.NewForConfig(kubeConfig)
	if err != nil {
		return nil, err
	}

	eventClient := clientSet.CoreV1().Events(kubeapi.NamespaceAll)
	eventSource := &EventSource{
		EventClient: eventClient,
		StopChannel: make(chan struct{}),
		EventBuffer: make(chan *kubeapi.Event, EventBufferSize),
	}

	go eventSource.Watch()

	return eventSource, nil
}
