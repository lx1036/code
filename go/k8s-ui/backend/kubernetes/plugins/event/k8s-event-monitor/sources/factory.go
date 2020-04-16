package sources

import (
	"errors"
	"fmt"
	"k8s-lx1036/k8s-ui/backend/kubernetes/plugins/event/k8s-event-monitor/common/flags"
	"k8s-lx1036/k8s-ui/backend/kubernetes/plugins/event/k8s-event-monitor/sources/kubernetes"
	"k8s.io/klog"
)

const (
	SrcKubernetes = "kubernetes"
)

type SourceFactory struct {
}

func NewSourceFactory() *SourceFactory {
	return &SourceFactory{}
}


func (factory *SourceFactory) Build(sources flags.Uris) (*kubernetes.EventSource, error) {
	var eventSource *kubernetes.EventSource
	var err error
	for _, source := range sources {
		switch source.Key {
		case SrcKubernetes:
			eventSource, err = kubernetes.NewKubernetesEventSource(&source.Value)
			if err != nil {
				return nil, err
			}
		default:
			klog.Errorf("Source[%s] is not supported.", source.Key)
			return nil, errors.New(fmt.Sprintf("Source[%s] is not supported.", source.Key))
		}
	}

	return eventSource, nil



	/*srcs := strings.Split(sources, ",")
	kubernetesSource := srcs[0]
	var uri = url.URL{}
	kubernetesUri, _ := uri.Parse(kubernetesSource)
	source, _ := kubernetes.NewKubernetesEventSource(kubernetesUri)
	return source*/
}
