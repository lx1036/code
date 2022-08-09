package podgroups

import (
	"context"
	"fmt"
	"time"

	podgroupv1 "k8s-lx1036/k8s/scheduler/pkg/apis/podgroup/v1"
	"k8s-lx1036/k8s/scheduler/pkg/client/clientset/versioned"
	"k8s-lx1036/k8s/scheduler/pkg/client/informers/externalversions"
	"k8s-lx1036/k8s/scheduler/pkg/framework"
	frameworkruntime "k8s-lx1036/k8s/scheduler/pkg/framework/runtime"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/cache"
	corev1helpers "k8s.io/component-helpers/scheduling/corev1"
	"k8s.io/klog/v2"
)

const (
	Name = "PodGroup"
)

type Coscheduling struct { // schedule pods in group
	framework       *frameworkruntime.Framework
	podGroupManager *PodGroupManager
	scheduleTimeout time.Duration
}

func New(args runtime.Object, framework *frameworkruntime.Framework) (framework.Plugin, error) {
	coschedulingArgs, ok := args.(*podgroupv1.CoschedulingArgs)
	if !ok {
		return nil, fmt.Errorf("want args to be of type CoschedulingArgs, got %v", args)
	}

	pgClient := versioned.NewForConfigOrDie(framework.KubeConfig())
	pgInformerFactory := externalversions.NewSharedInformerFactory(pgClient, 0)
	pgInformer := pgInformerFactory.PodGroup().V1().PodGroups()
	podInformer := framework.SharedInformerFactory().Core().V1().Pods()
	ctx := context.TODO()
	pgInformerFactory.Start(ctx.Done())
	if !cache.WaitForCacheSync(ctx.Done(), pgInformer.Informer().HasSynced) {
		err := fmt.Errorf("WaitForCacheSync failed")
		klog.ErrorS(err, "Cannot sync caches")
		return nil, err
	}

	scheduleTimeDuration := time.Duration(coschedulingArgs.PermitWaitingTimeSeconds) * time.Second
	pgMgr := NewPodGroupManager(pgClient, framework.SnapshotSharedLister(), scheduleTimeDuration, pgInformer, podInformer)
	plugin := &Coscheduling{
		framework:       framework,
		podGroupManager: pgMgr,
		scheduleTimeout: scheduleTimeDuration,
	}

	return plugin, nil
}

func (pl *Coscheduling) Name() string {
	return Name
}

func (pl *Coscheduling) Less(podInfo1, podInfo2 *framework.QueuedPodInfo) bool {
	prio1 := corev1helpers.PodPriority(podInfo1.Pod)
	prio2 := corev1helpers.PodPriority(podInfo2.Pod)
	if prio1 != prio2 {
		return prio1 > prio2
	}

	creationTime1 := pl.podGroupManager.GetCreationTimestamp(podInfo1.Pod, podInfo1.InitialAttemptTimestamp)
	creationTime2 := pl.podGroupManager.GetCreationTimestamp(podInfo2.Pod, podInfo2.InitialAttemptTimestamp)
	if creationTime1.Equal(creationTime2) {
		return GetNamespacedName(podInfo1.Pod) < GetNamespacedName(podInfo2.Pod)
	}

	return creationTime1.Before(creationTime2)
}

func GetNamespacedName(obj metav1.Object) string {
	return fmt.Sprintf("%v/%v", obj.GetNamespace(), obj.GetName())
}

// PreFilter
// 1. pod 不属于 pod-group，过滤掉
// 2. pods 不满足 pod-group MinMember 或者 MinResources，过滤掉
// @see https://github.com/kubernetes-sigs/scheduler-plugins/blob/master/kep/42-podgroup-coscheduling/README.md#prefilter
func (pl *Coscheduling) PreFilter(ctx context.Context, state *framework.CycleState,
	pod *corev1.Pod) (*framework.PreFilterResult, *framework.Status) {
	// If PreFilter fails, return framework.UnschedulableAndUnresolvable to avoid
	// any preemption attempts.
	if err := pl.podGroupManager.PreFilter(ctx, pod); err != nil {
		klog.ErrorS(err, "PreFilter failed", "pod", klog.KObj(pod))
		return nil, framework.NewStatus(framework.UnschedulableAndUnresolvable, err.Error())
	}
	return nil, framework.NewStatus(framework.Success, "")
}

func (pl *Coscheduling) PreFilterExtensions() framework.PreFilterExtensions {
	return nil
}

// PostFilter is used to reject a group of pods if a pod does not pass PreFilter or Filter.
func (pl *Coscheduling) PostFilter(ctx context.Context, state *framework.CycleState, pod *corev1.Pod,
	filteredNodeStatusMap framework.NodeToStatusMap) (*framework.PostFilterResult, *framework.Status) {
	pgName, pg := pl.podGroupManager.GetPodGroup(pod)
	if pg == nil {
		klog.V(4).InfoS("Pod does not belong to any group", "pod", klog.KObj(pod))
		return &framework.PostFilterResult{}, framework.NewStatus(framework.Unschedulable, "can not find pod group")
	}

	assigned := pl.podGroupManager.CalculateAssignedPods(pg.Name, pod.Namespace)
	if assigned >= int(pg.Spec.MinMember) {
		klog.V(4).InfoS("Assigned pods", "podGroup", klog.KObj(pg), "assigned", assigned)
		return &framework.PostFilterResult{}, framework.NewStatus(framework.Unschedulable)
	}
	// If the gap is less than/equal 10%, we may want to try subsequent Pods
	// to see they can satisfy the PodGroup
	notAssignedPercentage := float32(int(pg.Spec.MinMember)-assigned) / float32(pg.Spec.MinMember)
	if notAssignedPercentage <= 0.1 {
		klog.V(4).InfoS("A small gap of pods to reach the quorum", "podGroup", klog.KObj(pg), "percentage", notAssignedPercentage)
		return &framework.PostFilterResult{}, framework.NewStatus(framework.Unschedulable)
	}

	// INFO: 依次 reject test1 pod-group 里的 pod，这里很重要，是实现调度一组 pod 的核心逻辑!!!
	//  WaitOnPermit() 参考 PodGroup plugin，如果一组 pod 内，有一个 pod 调度失败了，
	//  会在 PostFilter 里逐个把其他 pod 给 reject 掉，因为其他 pod 这时已经在 WaitOnPermit()
	pl.framework.IterateOverWaitingPods(func(waitingPod *frameworkruntime.WaitingPod) {
		if waitingPod.GetPod().Namespace == pod.Namespace && waitingPod.GetPod().Labels[podgroupv1.PodGroupLabel] == pg.Name {
			klog.V(3).InfoS("PostFilter rejects the pod", "podGroup", klog.KObj(pg), "pod", klog.KObj(waitingPod.GetPod()))
			waitingPod.Reject(pl.Name(), "optimistic rejection in PostFilter")
		}
	})

	pl.podGroupManager.DeletePermittedPodGroup(pgName)

	return &framework.PostFilterResult{}, framework.NewStatus(framework.Unschedulable,
		fmt.Sprintf("PodGroup %v gets rejected due to Pod %v is unschedulable even after PostFilter", pgName, pod.Name))
}

func (pl *Coscheduling) Reserve(ctx context.Context, state *framework.CycleState, pod *corev1.Pod, nodeName string) *framework.Status {
	return nil
}

// Unreserve rejects all other Pods in the PodGroup when one of the pods in the group times out.
func (pl *Coscheduling) Unreserve(ctx context.Context, state *framework.CycleState, pod *corev1.Pod, nodeName string) {

}

func (pl *Coscheduling) Permit(ctx context.Context, state *framework.CycleState,
	pod *corev1.Pod, nodeName string) (*framework.Status, time.Duration) {

	s := pl.podGroupManager.Permit(ctx, pod)

}

// PostBind is called after a pod is successfully bound. These plugins are used update PodGroup when pod is bound.
func (pl *Coscheduling) PostBind(ctx context.Context, _ *framework.CycleState, pod *corev1.Pod, nodeName string) {

}
