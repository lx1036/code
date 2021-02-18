package sources

import (
	"errors"
	"fmt"
	"k8s.io/klog/v2"
)

const (
	SrcKubernetes = "kubernetes"
)

type SourceFactory struct {
}

func NewSourceFactory() *SourceFactory {
	return &SourceFactory{}
}

func (factory *SourceFactory) Build(sources flags.Uris) ([]common.EventSource, error) {
	var eventSources []common.EventSource
	for _, source := range sources {
		switch source.Key {
		case SrcKubernetes:
			eventSource, err := kubernetes.NewKubernetesEventSource(&source.Value)
			if err != nil {
				return nil, err
			}
			eventSources = append(eventSources, eventSource)
		default:
			klog.Errorf("Source[%s] is not supported.", source.Key)
			return nil, errors.New(fmt.Sprintf("Source[%s] is not supported.", source.Key))
		}
	}

	return eventSources, nil
}
