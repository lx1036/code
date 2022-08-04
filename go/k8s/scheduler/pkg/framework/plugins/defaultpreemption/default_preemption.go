package defaultpreemption

import (
	"context"
	"fmt"

	configv1 "k8s-lx1036/k8s/scheduler/pkg/apis/config/v1"
	"k8s-lx1036/k8s/scheduler/pkg/framework"
	frameworkruntime "k8s-lx1036/k8s/scheduler/pkg/framework/runtime"
	"k8s-lx1036/k8s/scheduler/pkg/metrics"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	corelisters "k8s.io/client-go/listers/core/v1"
	corev1helpers "k8s.io/component-helpers/scheduling/corev1"
	"k8s.io/klog/v2"
)

// INFO: 抢占preemption plugin

const (
	// Name of the plugin used in the plugin registry and configurations.
	Name = "DefaultPreemption"
)

// DefaultPreemption is a PostFilter plugin implements the preemption logic.
type DefaultPreemption struct {
	fh        framework.FrameworkHandle
	args      configv1.DefaultPreemptionArgs
	podLister corelisters.PodLister
}

func New(dpArgs runtime.Object, fh *frameworkruntime.Framework) (framework.Plugin, error) {
	args, ok := dpArgs.(*configv1.DefaultPreemptionArgs)
	if !ok {
		return nil, fmt.Errorf("got args of type %T, want *DefaultPreemptionArgs", dpArgs)
	}
	pl := DefaultPreemption{
		fh:        fh,
		args:      *args,
		podLister: fh.SharedInformerFactory().Core().V1().Pods().Lister(),
	}

	return &pl, nil
}

// Name returns name of the plugin. It is used in logs, etc.
func (pl *DefaultPreemption) Name() string {
	return Name
}

func (pl *DefaultPreemption) PostFilter(ctx context.Context, state *framework.CycleState, pod *v1.Pod,
	m framework.NodeToStatusMap) (*framework.PostFilterResult, *framework.Status) {
	defer func() {
		metrics.PreemptionAttempts.Inc()
	}()

	nominatedNodeName, err := pl.preempt(ctx, state, pod, m)
	if err != nil {
		return nil, framework.NewStatus(framework.Error, err.Error())
	}
	if nominatedNodeName == "" {
		return nil, framework.NewStatus(framework.Unschedulable)
	}

	return &framework.PostFilterResult{NominatedNodeName: nominatedNodeName}, framework.NewStatus(framework.Success)
}

func (pl *DefaultPreemption) preempt(ctx context.Context, state *framework.CycleState, pod *v1.Pod,
	m framework.NodeToStatusMap) (string, error) {
	nodeLister := pl.fh.SnapshotSharedLister().NodeInfos()

	podNamespace, podName := pod.Namespace, pod.Name
	pod, err := pl.podLister.Pods(pod.Namespace).Get(pod.Name)
	if err != nil {
		klog.ErrorS(err, "Getting the updated preemptor pod object", "pod", klog.KRef(podNamespace, podName))
		return nil, framework.AsStatus(err)
	}

	// INFO:(1) Ensure the preemptor is eligible to preempt other pods.
	if !pl.PodEligibleToPreemptOthers(pod, m[pod.Status.NominatedNodeName]) {
		klog.V(5).Infof("Pod %v/%v is not eligible for more preemption.", pod.Namespace, pod.Name)
		return "", nil
	}

	// INFO:(2) Find all preemption candidates.
	candidates, nodeToStatusMap, err := pl.findCandidates(ctx, pod, m)
	if err != nil && len(candidates) == 0 {
		return nil, framework.AsStatus(err)
	}

	// INFO: (3) Find the best candidate.
	bestCandidate := pl.SelectCandidate(candidates)
	if bestCandidate == nil || len(bestCandidate.Name()) == 0 {
		return nil, framework.NewStatus(framework.Unschedulable, "no candidate node for preemption")
	}

	// INFO: (4) Perform preparation work before nominating the selected candidate.
	if status := pl.prepareCandidate(bestCandidate, pod, pl.PluginName); !status.IsSuccess() {
		return nil, status
	}

}

// PodEligibleToPreemptOthers INFO: @see https://kubernetes.io/zh/docs/concepts/configuration/pod-priority-preemption/#non-preempting-priority-class
func (pl *DefaultPreemption) PodEligibleToPreemptOthers(pod *v1.Pod, nominatedNodeStatus *framework.Status) bool {
	// INFO: 非抢占式的 pod 不需要抢占，返回 false
	if pod.Spec.PreemptionPolicy != nil && *pod.Spec.PreemptionPolicy == v1.PreemptNever {
		klog.V(5).Infof("Pod %v/%v is not eligible for preemption because it has a preemptionPolicy of %v", pod.Namespace, pod.Name, v1.PreemptNever)
		return false
	}

	nodeInfos := pl.fh.SnapshotSharedLister().NodeInfos()
	nominatedNodeName := pod.Status.NominatedNodeName // 如果之前已经抢占过了，则检查下
	if len(nominatedNodeName) > 0 {
		// INFO: nominatedNodeName 是不可调度的，可以抢占？
		if nominatedNodeStatus.Code() == framework.UnschedulableAndUnresolvable {
			return true
		}

		// nominatedNodeName 有 terminating pod 了，则不需要抢占
		if nodeInfo, _ := nodeInfos.Get(nominatedNodeName); nodeInfo != nil {
			podPriority := corev1helpers.PodPriority(pod)
			for _, p := range nodeInfo.Pods {
				if p.Pod.DeletionTimestamp != nil && corev1helpers.PodPriority(p.Pod) < podPriority {
					// There is a terminating pod on the nominated node.
					return false
				}
			}
		}
	}

	return true
}
