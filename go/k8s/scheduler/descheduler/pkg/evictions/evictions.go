package evictions

import (
	"fmt"

	podutil "k8s-lx1036/k8s/scheduler/descheduler/pkg/pod"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/errors"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
)

const (
	EvictionKind        = "Eviction"
	EvictionSubresource = "pods/eviction"
)

type PodEvictor struct {
	client                clientset.Interface
	policyGroupVersion    string
	dryRun                bool
	maxPodsToEvictPerNode int
	nodepodCount          nodePodEvictedCount
	evictLocalStoragePods bool
}

// nodePodEvictedCount keeps count of pods evicted on node
type nodePodEvictedCount map[*v1.Node]int // node上驱逐pod的数量

func NewPodEvictor(
	client clientset.Interface,
	policyGroupVersion string,
	dryRun bool,
	maxPodsToEvictPerNode int,
	nodes []*v1.Node,
	evictLocalStoragePods bool,
) *PodEvictor {
	var nodePodCount = make(nodePodEvictedCount)
	for _, node := range nodes {
		nodePodCount[node] = 0
	}

	return &PodEvictor{
		client:                client,
		policyGroupVersion:    policyGroupVersion,
		dryRun:                dryRun,
		maxPodsToEvictPerNode: maxPodsToEvictPerNode,
		nodepodCount:          nodePodCount,
		evictLocalStoragePods: evictLocalStoragePods,
	}
}

type constraint func(pod *v1.Pod) error
type evictable struct {
	constraints []constraint
}

// Evictable provides an implementation of IsEvictable(IsEvictable(pod *v1.Pod) bool).
// The method accepts a list of options which allow to extend constraints
// which decides when a pod is considered evictable.
func (pe *PodEvictor) Evictable(opts ...func(opts *Options)) *evictable {
	options := &Options{}
	for _, opt := range opts {
		opt(options)
	}

	ev := &evictable{}
	if !pe.evictLocalStoragePods {
		ev.constraints = append(ev.constraints, func(pod *v1.Pod) error {
			if IsPodWithLocalStorage(pod) {
				return fmt.Errorf("pod has local storage and descheduler is not configured with --evict-local-storage-pods")
			}
			return nil
		})
	}

	if options.priority != nil {
		ev.constraints = append(ev.constraints, func(pod *v1.Pod) error {
			// 小于 priority threshold 的pod都要被驱逐evict
			if IsPodEvictableBasedOnPriority(pod, *options.priority) {
				return nil
			}
			return fmt.Errorf("pod has higher priority than specified priority class threshold")
		})
	}

	return ev
}

type Options struct {
	priority *int32
}

// WithPriorityThreshold sets a threshold for pod's priority class.
// Any pod whose priority class is lower is evictable.
func WithPriorityThreshold(priority int32) func(opts *Options) {
	return func(opts *Options) {
		var p int32 = priority
		opts.priority = &p
	}
}

// 小于 priority threshold 的pod都要被驱逐evict
func IsPodEvictableBasedOnPriority(pod *v1.Pod, priority int32) bool {
	return pod.Spec.Priority == nil || *pod.Spec.Priority < priority
}

// 有 local storage的pod，即hostPath和EmptyDir
func IsPodWithLocalStorage(pod *v1.Pod) bool {
	for _, volume := range pod.Spec.Volumes {
		if volume.HostPath != nil || volume.EmptyDir != nil {
			return true
		}
	}

	return false
}

// SupportEviction uses Discovery API to find out if the server support eviction subresource
// If support, it will return its groupVersion; Otherwise, it will return ""
func SupportEviction(client clientset.Interface) (string, error) {
	discoveryClient := client.Discovery()
	groupList, err := discoveryClient.ServerGroups()
	if err != nil {
		return "", err
	}

	//klog.Infof("groupList.Groups: %v", groupList.Groups)
	foundPolicyGroup := false
	var policyGroupVersion string
	for _, group := range groupList.Groups {
		if group.Name == "policy" {
			klog.Infof("[SupportEviction]group policy: %v", group)
			foundPolicyGroup = true
			policyGroupVersion = group.PreferredVersion.GroupVersion
			break
		}
	}
	if !foundPolicyGroup {
		return "", nil
	}
	resourceList, err := discoveryClient.ServerResourcesForGroupVersion("v1")
	if err != nil {
		return "", err
	}

	//klog.Infof("resourceList.APIResources: %v", resourceList.APIResources)
	for _, resource := range resourceList.APIResources {
		if resource.Name == EvictionSubresource && resource.Kind == EvictionKind {
			klog.Infof("[SupportEviction]resource: %v", resource)

			return policyGroupVersion, nil
		}
	}
	return "", nil
}

// critical pod, daemonset pod, mirror pod 不驱逐
func (ev *evictable) IsEvictable(pod *v1.Pod) bool {
	var checkErrs []error

	// 不驱逐 static pod，mirror pod 和 高优先级system-cluster-critical pod
	if podutil.IsCriticalPod(pod) {
		checkErrs = append(checkErrs, fmt.Errorf("pod is critical"))
	}

	ownerRefList := pod.ObjectMeta.GetOwnerReferences()
	if podutil.IsDaemonsetPod(ownerRefList) {
		checkErrs = append(checkErrs, fmt.Errorf("pod is a DaemonSet pod"))
	}

	// 无主pod
	if len(ownerRefList) == 0 {
		checkErrs = append(checkErrs, fmt.Errorf("pod does not have any ownerrefs"))
	}

	// 经过ev.constraints check之后
	for _, c := range ev.constraints {
		if err := c(pod); err != nil {
			checkErrs = append(checkErrs, err)
		}
	}

	if len(checkErrs) > 0 && !HaveEvictAnnotation(pod) { // 根据标记判断，之前没有被驱逐过
		klog.V(4).InfoS("Pod lacks an eviction annotation and fails the following checks", "pod", klog.KObj(pod), "checks", errors.NewAggregate(checkErrs).Error())
		return false
	}

	return true
}

const (
	evictPodAnnotationKey = "descheduler.alpha.kubernetes.io/evict"
)

// pod 已经被驱逐过了
func HaveEvictAnnotation(pod *v1.Pod) bool {
	_, found := pod.ObjectMeta.Annotations[evictPodAnnotationKey]
	return found
}
