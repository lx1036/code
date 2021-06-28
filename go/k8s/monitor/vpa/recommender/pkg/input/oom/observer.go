package oom

import (
	"context"
	"strings"
	"time"

	"k8s-lx1036/k8s/monitor/vpa/recommender/pkg/controller/clusterstate/types"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/kubelet/eviction"
)

// INFO: kubelet 在驱逐 pod 时，会产生一个 event https://github.com/kubernetes/kubernetes/blob/v1.19.7/pkg/kubelet/eviction/eviction_manager.go#L583
// 并把 corev1.Event 对象写入 apiserver 中，这个 observer 对象会 watch 这个 Event 对象，并写入到 observedOomsChannel 中

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
	if event.Reason != eviction.Reason || event.InvolvedObject.Kind != "Pod" {
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
	/*
		annotations[OffendingContainersKey] = strings.Join(containers, ",")
		annotations[OffendingContainersUsageKey] = strings.Join(containerUsage, ",")
		annotations[StarvedResourceKey] = string(resourceToReclaim)
	*/
	offendingContainers := extractArray(eviction.OffendingContainersKey)
	offendingContainersUsage := extractArray(eviction.OffendingContainersUsageKey)
	starvedResource := extractArray(eviction.StarvedResourceKey)
	if len(offendingContainers) != len(offendingContainersUsage) || len(offendingContainers) != len(starvedResource) {
		return []OomInfo{}
	}

	result := make([]OomInfo, 0, len(offendingContainers))
	for i, containerName := range offendingContainers {
		if starvedResource[i] != string(corev1.ResourceMemory) {
			continue
		}
		// INFO: 这里直接通过 i 来获取 container usage，是因为 kubelet 代码里 https://github.com/kubernetes/kubernetes/blob/v1.19.7/pkg/kubelet/eviction/helpers.go#L1065-L1069
		memory, err := resource.ParseQuantity(offendingContainersUsage[i])
		if err != nil {
			klog.Errorf("Cannot parse resource quantity in eviction event %v. Error: %v", offendingContainersUsage[i], err)
			continue
		}
		oomInfo := OomInfo{
			Timestamp: event.CreationTimestamp.Time.UTC(),
			Memory:    types.ResourceAmount(memory.Value()),
			ContainerID: types.ContainerID{
				PodID: types.PodID{
					Namespace: event.InvolvedObject.Namespace,
					PodName:   event.InvolvedObject.Name,
				},
				ContainerName: containerName,
			},
		}
		result = append(result, oomInfo)
	}

	return result
}

const (
	evictionWatchRetryWait    = 10 * time.Second
	evictionWatchJitterFactor = 0.5
)

// WatchEvictionEventsWithRetries INFO: 这个逻辑可以复用：watch event 事件，如果 watch 过程中有失败，则等待 jitter 时间重新去 watch
func WatchEvictionEventsWithRetries(kubeClient kubernetes.Interface, observer Observer, namespace string, stopCh <-chan struct{}) {
	go func() {
		options := metav1.ListOptions{
			FieldSelector: "reason=Evicted",
		}

		watchEvictionEventsOnce := func() {
			watcher, err := kubeClient.CoreV1().Events(namespace).Watch(context.TODO(), options)
			if err != nil {
				klog.Errorf("Cannot initialize watching events. Reason %v", err)
				return
			}

			// 读取 event 数据
			for {
				evictedEvent, ok := <-watcher.ResultChan() // 这里 channel 阻塞，会一直 watch 新的 event 到来
				if !ok {
					klog.V(3).Infof("Eviction event chan closed")
					return
				}
				event, ok := evictedEvent.Object.(*corev1.Event)
				if !ok {
					return
				}

				observer.OnEvent(event)
			}
		}

		for {
			watchEvictionEventsOnce()
			// Wait between attempts, retrying too often breaks API server.
			waitTime := wait.Jitter(evictionWatchRetryWait, evictionWatchJitterFactor)
			klog.V(1).Infof("An attempt to watch eviction events finished. Waiting %v before the next one.", waitTime)
			time.Sleep(waitTime)
		}
	}()
}
