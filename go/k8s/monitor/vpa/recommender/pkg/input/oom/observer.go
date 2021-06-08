package oom

import (
	"context"
	"strings"
	"time"

	"k8s-lx1036/k8s/kubelet/pkg/eviction"
	"k8s-lx1036/k8s/monitor/vpa/recommender/pkg/types"
	corev1 "k8s.io/api/core/v1"
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

func (o *Observer) OnEvent(event *corev1.Event) {
	klog.V(1).Infof("OOM Observer processing event: %+v", event)
	for _, oomInfo := range parseEvictionEvent(event) {
		o.observedOomsChannel <- oomInfo
	}
}

func parseEvictionEvent(event *corev1.Event) []OomInfo {
	// 必须是 Pod 产生的 Evicted 事件
	if event.Reason != "Evicted" || event.InvolvedObject.Kind != "Pod" {
		return []OomInfo{}
	}

	// INFO: 这里有个知识点，kubelet evict pod 时会产生 Evicted event, 见代码：https://github.com/kubernetes/kubernetes/blob/v1.19.7/pkg/kubelet/eviction/eviction_manager.go#L569-L592
	// INFO: 同时 event 带有这三个 annotation，见代码：https://github.com/kubernetes/kubernetes/blob/v1.19.7/pkg/kubelet/eviction/helpers.go#L1040-L1077
	extractArray := func(annotationsKey string) []string {
		str, found := event.Annotations[annotationsKey]
		if !found {
			return []string{}
		}
		return strings.Split(str, ",")
	}
	// 三个字段值数量必须相等
	offendingContainers := extractArray(eviction.OffendingContainersKey)
	offendingContainersUsage := extractArray(eviction.OffendingContainersUsageKey)
	starvedResource := extractArray(eviction.StarvedResourceKey)
	if len(offendingContainers) != len(offendingContainersUsage) ||
		len(offendingContainers) != len(starvedResource) {
		return []OomInfo{}
	}

}

// WatchEvictionEventsWithRetries INFO: 这个逻辑可以复用：watch event 事件，如果 watch 过程中有失败，则等待 jitter 时间重新去 watch
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

			if evictedEvent.Type == watch.Added { // INFO: 只有新建的 event
				evictedEvent, ok := evictedEvent.Object.(*corev1.Event)
				if !ok {
					continue
				}
				observer.OnEvent(evictedEvent)
			}
		}
	}, time.Second*10, stopCh)
}
