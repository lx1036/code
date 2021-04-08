package pod

import (
	"context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
	"k8s.io/component-base/featuregate"

	v1 "k8s.io/api/core/v1"
	clientset "k8s.io/client-go/kubernetes"
)

// Enables PodOverhead, for accounting pod overheads which are specific to a given RuntimeClass
const PodOverhead featuregate.Feature = "PodOverhead"

type Options struct {
	filter             func(pod *v1.Pod) bool
	includedNamespaces []string
	excludedNamespaces []string
}

// TODO: list node上的pods，同时带有过滤功能，可以直接复用。注意，这里是直接从apiserver取值，没有
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
	// TODO: field selectors do not work properly with listers
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

// TODO: 计算pod 各种资源总和，可以直接复用
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
	if pod.Spec.Overhead != nil && utilfeature.DefaultFeatureGate.Enabled(PodOverhead) {
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
