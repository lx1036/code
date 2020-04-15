package sources

import (
	"k8s-lx1036/k8s-ui/backend/kubernetes/plugins/event/k8s-event-monitor/sources/kubernetes"
	"net/url"
	"strings"
)

type SourceFactory struct {
}

func NewSourceFactory() *SourceFactory {
	return &SourceFactory{}
}

func (factory *SourceFactory) BuildAll(sources string) *kubernetes.EventSource {
	srcs := strings.Split(sources, ",")
	kubernetesSource := srcs[0]
	var uri = url.URL{}
	kubernetesUri, _ := uri.Parse(kubernetesSource)
	source, _ := kubernetes.NewKubernetesEventSource(kubernetesUri)
	return source
}
