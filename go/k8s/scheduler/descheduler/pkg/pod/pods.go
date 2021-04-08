package pod

import (
	"context"
	"fmt"
	"sort"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/apis/core/v1/helper"
	v1qos "k8s.io/kubernetes/pkg/apis/core/v1/helper/qos"
	"k8s.io/kubernetes/pkg/apis/scheduling"
	"k8s.io/kubernetes/pkg/features"
	"k8s.io/kubernetes/pkg/kubelet/types"
)

type Options struct {
	filter             func(pod *v1.Pod) bool
	includedNamespaces []string
	excludedNamespaces []string
}

// INFO: list node上的pods，同时带有过滤功能，可以直接复用。注意，这里是直接从apiserver取值，没有
// Usually this is podEvictor.Evictable().IsEvictable, 可以用来list the evictable pods on a node
func ListPodsOnNode(
	ctx context.Context,
	client clientset.Interface,
	node *v1.Node,
	opts ...func(opts *Options),
) ([]*v1.Pod, error) {
	options := &Options{}
	for _, opt := range opts {
		opt(options)
	}

	pods := make([]*v1.Pod, 0)

	// 只会驱逐 Running/Pending/Unknown pods
	fieldSelectorString := "spec.nodeName=" + node.Name + ",status.phase!=" + string(v1.PodSucceeded) + ",status.phase!=" + string(v1.PodFailed)
	// 只考虑该node上的该namespaces下的pods
	if len(options.includedNamespaces) > 0 {
		fieldSelector, err := fields.ParseSelector(fieldSelectorString)
		if err != nil {
			return []*v1.Pod{}, err
		}

		for _, namespace := range options.includedNamespaces {
			// 从 apiserver 中取pods
			podList, err := client.CoreV1().Pods(namespace).List(ctx,
				metav1.ListOptions{FieldSelector: fieldSelector.String()})
			if err != nil {
				return []*v1.Pod{}, err
			}
			for i := range podList.Items {
				if options.filter != nil && !options.filter(&podList.Items[i]) {
					continue
				}
				pods = append(pods, &podList.Items[i])
			}
		}
		return pods, nil
	}

	if len(options.excludedNamespaces) > 0 {
		for _, namespace := range options.excludedNamespaces {
			fieldSelectorString += ",metadata.namespace!=" + namespace
		}
	}
	fieldSelector, err := fields.ParseSelector(fieldSelectorString)
	if err != nil {
		return []*v1.Pod{}, err
	}
	// 从 apiserver 中取pods
	// INFO: field selectors do not work properly with listers
	podList, err := client.CoreV1().Pods(v1.NamespaceAll).List(ctx,
		metav1.ListOptions{FieldSelector: fieldSelector.String()})
	if err != nil {
		return []*v1.Pod{}, err
	}
	for i := range podList.Items {
		// fake client does not support field selectors
		// so let's filter based on the node name as well (quite cheap)
		if podList.Items[i].Spec.NodeName != node.Name {
			continue
		}
		if options.filter != nil && !options.filter(&podList.Items[i]) {
			continue
		}
		pods = append(pods, &podList.Items[i])
	}

	return pods, nil
}

// INFO: 计算pod 各种资源request/limit总和，可以直接复用
func PodRequestsAndLimits(pod *v1.Pod) (reqs, limits v1.ResourceList) {
	reqs, limits = v1.ResourceList{}, v1.ResourceList{}
	// addResourceList依次累加到reqs, limits
	for _, container := range pod.Spec.Containers {
		addResourceList(reqs, container.Resources.Requests)
		addResourceList(limits, container.Resources.Limits)
	}
	for _, container := range pod.Spec.InitContainers {
		maxResourceList(reqs, container.Resources.Requests)
		maxResourceList(limits, container.Resources.Limits)
	}

	// 如果有PodOverhead继续累加
	if pod.Spec.Overhead != nil && utilfeature.DefaultFeatureGate.Enabled(features.PodOverhead) {
		addResourceList(reqs, pod.Spec.Overhead)

		for name, quantity := range pod.Spec.Overhead {
			if value, ok := limits[name]; ok && !value.IsZero() {
				value.Add(quantity)
				limits[name] = value
			}
		}
	}

	return
}

// 依次累加newList
func addResourceList(list, newList v1.ResourceList) {
	for name, quantity := range newList {
		if value, ok := list[name]; !ok {
			list[name] = quantity.DeepCopy() // 这里防止影响原quantity最好直接复制个新对象
		} else {
			value.Add(quantity)
			list[name] = value
		}
	}
}

// InitContainers和Containers同一资源用最大值
func maxResourceList(list, new v1.ResourceList) {
	for name, quantity := range new {
		if value, ok := list[name]; !ok {
			list[name] = quantity.DeepCopy()
			continue
		} else {
			if quantity.Cmp(value) > 0 {
				list[name] = quantity.DeepCopy()
			}
		}
	}
}

// 根据 pod 优先级升序排序，优先级相等则根据QoS(BestEffort, Burstable, Guaranteed)升序排序
func SortPodsBasedOnPriorityLowToHigh(pods []*v1.Pod) {
	sort.Slice(pods, func(i, j int) bool {
		// i 没有 Priority, j 有，则 i 在前
		if pods[i].Spec.Priority == nil && pods[j].Spec.Priority != nil {
			return true
		}
		// j 没有 Priority, i 有，则 j 在前
		if pods[j].Spec.Priority == nil && pods[i].Spec.Priority != nil {
			return false
		}
		// 都没有 Priority 或者相等
		if (pods[j].Spec.Priority == nil && pods[i].Spec.Priority == nil) ||
			(*pods[i].Spec.Priority == *pods[j].Spec.Priority) {
			// i 是 BestEffort pod，i 在前
			if IsBestEffortPod(pods[i]) {
				return true
			}
			// i 是 Burstable 且 j 是 Guaranteed，i 在前
			if IsBurstablePod(pods[i]) && IsGuaranteedPod(pods[j]) {
				return true
			}
			// j 在前
			return false
		}

		// 谁 Priority 优先级小谁在前
		return *pods[i].Spec.Priority < *pods[j].Spec.Priority
	})
}

func IsBestEffortPod(pod *v1.Pod) bool {
	return v1qos.GetPodQOS(pod) == v1.PodQOSBestEffort
}

func IsBurstablePod(pod *v1.Pod) bool {
	return v1qos.GetPodQOS(pod) == v1.PodQOSBurstable
}

func IsGuaranteedPod(pod *v1.Pod) bool {
	return v1qos.GetPodQOS(pod) == v1.PodQOSGuaranteed
}

// 不驱逐 static pod，mirror pod 和 高优先级system-cluster-critical pod
func IsCriticalPod(pod *v1.Pod) bool {
	if IsStaticPod(pod) {
		return true
	}

	if IsMirrorPod(pod) {
		return true
	}

	if pod.Spec.Priority != nil && *pod.Spec.Priority >= scheduling.SystemCriticalPriority {
		return true
	}

	return false
}

func IsStaticPod(pod *v1.Pod) bool {
	source, err := GetPodSource(pod)
	return err == nil && source != "api"
}

// GetPodSource returns the source of the pod based on the annotation.
func GetPodSource(pod *v1.Pod) (string, error) {
	if pod.Annotations != nil {
		if source, ok := pod.Annotations[types.ConfigSourceAnnotationKey]; ok {
			return source, nil
		}
	}
	return "", fmt.Errorf("cannot get source of pod %q", pod.UID)
}

func IsMirrorPod(pod *v1.Pod) bool {
	_, ok := pod.Annotations[v1.MirrorPodAnnotationKey]
	return ok
}

func IsDaemonsetPod(ownerRefList []metav1.OwnerReference) bool {
	for _, ownerRef := range ownerRefList {
		if ownerRef.Kind == "DaemonSet" {
			return true
		}
	}
	return false
}

// 如果pod能容忍node taints则返回true
func PodToleratesTaints(pod *v1.Pod, taintsOfNodes map[string][]v1.Taint) bool {
	for nodeName, taintsForNode := range taintsOfNodes {
		// 判断pod.Spec.Tolerations是否可以容忍[]v1.Taint
		if len(pod.Spec.Tolerations) >= len(taintsForNode) {
			_, isUntolerated := helper.FindMatchingUntoleratedTaint(taintsForNode, pod.Spec.Tolerations, nil)
			if !isUntolerated {
				return true
			}
		}

		klog.V(5).InfoS("Pod doesn't tolerate nodes taint", "pod", klog.KObj(pod), "nodeName", nodeName)
	}

	return false
}
