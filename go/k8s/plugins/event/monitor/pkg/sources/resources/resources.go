package resources

import (
	"fmt"
	"github.com/spf13/viper"
	"k8s.io/api/apps/v1beta2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"strconv"
)

type Kind string

const (
	Pod                   Kind = "pod"
	ReplicationController Kind = "replicationcontroller"
	ReplicaSet            Kind = "replicaset"
	DaemonSet             Kind = "daemonset"
	Service               Kind = "service"
	Deployment            Kind = "deployment"
	Node                  Kind = "node"
	Event                 Kind = "event"
	Ingress               Kind = "ingress"
	Secret                Kind = "secret"
	Configmap             Kind = "configmap"
)

var SupportedResources = []Kind{
	Pod,
	Node,
}

var Resources = map[Kind]schema.GroupVersionResource{
	Deployment: v1beta2.SchemeGroupVersion.WithResource("deployments"),
	Node:       corev1.SchemeGroupVersion.WithResource("nodes"),
	Event:      corev1.SchemeGroupVersion.WithResource("events"),
	Pod:        corev1.SchemeGroupVersion.WithResource("pods"),
	Ingress:    corev1.SchemeGroupVersion.WithResource("ingresses"),
}

func GetWatchedResources() ([]Kind, error) {
	var watchedResources []Kind
	resources := viper.Get("resources")
	if resources == nil {
		return nil, fmt.Errorf("resources field can't be empty")
	}

	for resource, opened := range resources.(map[Kind]string) {
		open, err := strconv.ParseBool(opened)
		if err != nil {
			continue
		}

		if open && Find(SupportedResources, resource) {
			watchedResources = append(watchedResources, resource)
		}
	}

	return watchedResources, nil
}

func Find(resources []Kind, resource Kind) bool {
	for _, value := range resources {
		if value == resource {
			return true
		}
	}

	return false
}
