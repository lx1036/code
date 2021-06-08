package oom

import (
	"context"
	"time"

	"k8s-lx1036/k8s/monitor/vpa/recommender/pkg/types"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
)

// INFO: watch pod oom event

type OomInfo struct {
	Timestamp   time.Time
	Memory      types.ResourceAmount
	ContainerID types.ContainerID
}

// observer can observe pod resource update and collect OOM events.
type Observer struct {
	observedOomsChannel chan OomInfo
}

func NewObserver() *Observer {
	return &Observer{
		observedOomsChannel: make(chan OomInfo, 5000),
	}
}

// WatchEvictionEventsWithRetries watches new Events with reason=Evicted and passes them to the observer.
func WatchEvictionEventsWithRetries(kubeClient kubernetes.Interface, observer Observer, namespace string, stopCh <-chan struct{}) {
	go wait.Until(func() {
		options := metav1.ListOptions{
			FieldSelector: "reason=Evicted",
		}
		watcher, err := kubeClient.CoreV1().Events(namespace).Watch(context.TODO(), options)
		if err != nil {
			klog.Errorf("Cannot initialize watching events. Reason %v", err)
			return
		}

		for {
			evictedEvent, ok := <-watcher.ResultChan()
			if !ok {
				klog.V(3).Infof("Eviction event chan closed")
				return
			}

			if evictedEvent.Type == watch.Added {
				evictedEvent, ok := evictedEvent.Object.(*apiv1.Event)
				if !ok {
					continue
				}
				observer.OnEvent(evictedEvent)
			}
		}
	}, time.Second*10, stopCh)
}
