package evictions

import (
	"context"
	"fmt"
	"strings"

	podutil "k8s-lx1036/k8s/scheduler/descheduler/pkg/pod"

	v1 "k8s.io/api/core/v1"
	policyv1beta1 "k8s.io/api/policy/v1beta1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/errors"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	clientcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
)

const (
	EvictionKind        = "Eviction"
	EvictionSubresource = "pods/eviction"
)

type PodEvictor struct {
	client                clientset.Interface
	policyGroupVersion    string
	dryRun                bool // 有 dryRun 特别方便本地调试
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

// TotalEvicted gives a number of pods evicted through all nodes
// 所有nodes
func (pe *PodEvictor) TotalEvicted() int {
	var total int
	for _, count := range pe.nodepodCount {
		total += count
	}
	return total
}

func (pe *PodEvictor) EvictPod(ctx context.Context, pod *v1.Pod, node *v1.Node, reasons ...string) (bool, error) {
	var reason string
	if len(reasons) > 0 {
		reason = " (" + strings.Join(reasons, ", ") + ")"
	}
	if pe.maxPodsToEvictPerNode > 0 && pe.nodepodCount[node]+1 > pe.maxPodsToEvictPerNode {
		return false, fmt.Errorf("maximum number %v of evicted pods per %q node reached",
			pe.maxPodsToEvictPerNode, node.Name)
	}

	err := evictPod(ctx, pe.client, pod, pe.policyGroupVersion, pe.dryRun)
	if err != nil {
		klog.ErrorS(err, "Error evicting pod", "pod", klog.KObj(pod), "reason", reason)
		return false, nil
	}

	// 该node上被驱逐pod数量
	pe.nodepodCount[node]++

	if pe.dryRun {
		klog.V(1).InfoS("Evicted pod in dry run mode", "pod", klog.KObj(pod), "reason", reason)
	} else {
		// 给 pod 添加个 event
		klog.V(1).InfoS("Evicted pod", "pod", klog.KObj(pod), "reason", reason)
		eventBroadcaster := record.NewBroadcaster()
		eventBroadcaster.StartStructuredLogging(3)
		eventBroadcaster.StartRecordingToSink(&clientcorev1.EventSinkImpl{Interface: pe.client.CoreV1().Events(pod.Namespace)})
		r := eventBroadcaster.NewRecorder(scheme.Scheme, v1.EventSource{Component: "sigs.k8s.io.descheduler"})
		r.Event(pod, v1.EventTypeNormal, "Descheduled", fmt.Sprintf("pod evicted by sigs.k8s.io/descheduler%s", reason))
	}

	return true, nil
}

// INFO: 驱逐pod实际上就是创建 pods/eviction 子资源而已
func evictPod(ctx context.Context, client clientset.Interface,
	pod *v1.Pod, policyGroupVersion string, dryRun bool) error {
	if dryRun {
		return nil
	}

	// INFO: apierrors 包还是很实用的，可以在调用 apiserver 返回error时，用起来
	eviction := &policyv1beta1.Eviction{
		TypeMeta: metav1.TypeMeta{
			APIVersion: policyGroupVersion,
			Kind:       EvictionKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      pod.Name,
			Namespace: pod.Namespace,
		},
		DeleteOptions: &metav1.DeleteOptions{},
	}
	err := client.PolicyV1beta1().Evictions(eviction.Namespace).Evict(ctx, eviction)
	if apierrors.IsTooManyRequests(err) {
		return fmt.Errorf("error when evicting pod (ignoring) %q: %v", pod.Name, err)
	}
	if apierrors.IsNotFound(err) {
		return fmt.Errorf("pod not found when evicting %q: %v", pod.Name, err)
	}
	return err
}

// node中已经被驱逐pod数量
func (pe *PodEvictor) NodeEvicted(node *v1.Node) int {
	return pe.nodepodCount[node]
}

type constraint func(pod *v1.Pod) error
type evictable struct {
	constraints []constraint
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

// 查找支持的 eviction GroupVersion，如 "policy/v1beta1"
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
			klog.Infof("[SupportEviction]group policy: %v, policyGroupVersion: %s",
				group, group.PreferredVersion.GroupVersion)
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

const (
	evictPodAnnotationKey = "descheduler.alpha.kubernetes.io/evict"
)

// pod 已经被驱逐过了
func HaveEvictAnnotation(pod *v1.Pod) bool {
	_, found := pod.ObjectMeta.Annotations[evictPodAnnotationKey]
	return found
}
