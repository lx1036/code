package kubernetes

import (
	"fmt"
	kubeapi "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubewatch "k8s.io/apimachinery/pkg/watch"
	kubeclient "k8s.io/client-go/kubernetes"
	kubev1core "k8s.io/client-go/kubernetes/typed/core/v1"
	kuberest "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
	"net/url"
	"strconv"
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
							fmt.Println(event.Message)
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

func (eventSource *EventSource) ListEvents() {

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

	configOverrides, err := getConfigOverrides(uri)
	if err != nil {
		return nil, err
	}

	query := uri.Query()
	inClusterConfig := defaultInClusterConfig
	if len(query["inClusterConfig"]) > 0 {
		inClusterConfig, err = strconv.ParseBool(query["inClusterConfig"][0])
		if err != nil {
			return nil, err
		}
	}
	if inClusterConfig {
		kubeConfig, err = kuberest.InClusterConfig()
		if err != nil {
			return nil, err
		}
	} else {
		authFile := ""
		if len(query["auth"]) > 0 {
			authFile = query["auth"][0]
		}
		if len(authFile) != 0 {

		} else {
			kubeConfig = &kuberest.Config{
				Host: configOverrides.ClusterInfo.Server,
				TLSClientConfig: kuberest.TLSClientConfig{
					Insecure: configOverrides.ClusterInfo.InsecureSkipTLSVerify,
				},
			}
		}
	}

	return kubeConfig, nil
}

func getConfigOverrides(uri *url.URL) (*clientcmd.ConfigOverrides, error) {
	configOverrides := &clientcmd.ConfigOverrides{
		ClusterInfo: api.Cluster{},
	}
	if len(uri.Host) != 0 && len(uri.Scheme) != 0 {
		configOverrides.ClusterInfo.Server = fmt.Sprintf("%s://%s", uri.Scheme, uri.Host)
	}
	query := uri.Query()
	if len(query["insecure"]) != 0 {
		insecure, err := strconv.ParseBool(query["insecure"][0])
		if err != nil {
			return nil, err
		}
		configOverrides.ClusterInfo.InsecureSkipTLSVerify = insecure
	}

	return configOverrides, nil
}
