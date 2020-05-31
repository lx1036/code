package common

import (
	"k8s-lx1036/k8s-ui/dashboard/controllers/resource/event"
	"k8s-lx1036/k8s-ui/dashboard/controllers/resource/pod"
)

type ResourceChannels struct {
	PodListChannel   pod.PodListChannel
	EventListChannel event.EventListChannel
}
