package defaultpreemption

import (
	"context"

	framework "k8s-lx1036/k8s/scheduler/pkg/scheduler/framework/v1alpha1"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	policylisters "k8s.io/client-go/listers/policy/v1beta1"
	"k8s.io/klog/v2"
	podutil "k8s.io/kubernetes/pkg/api/v1/pod"
)

// INFO: 抢占preemption plugin

const (
	// Name of the plugin used in the plugin registry and configurations.
	Name = "DefaultPreemption"
)

// DefaultPreemption is a PostFilter plugin implements the preemption logic.
type DefaultPreemption struct {
	fh        framework.FrameworkHandle
	pdbLister policylisters.PodDisruptionBudgetLister
}

// Name returns name of the plugin. It is used in logs, etc.
func (pl *DefaultPreemption) Name() string {
	return Name
}

func (pl *DefaultPreemption) PostFilter(ctx context.Context, state *framework.CycleState, pod *v1.Pod,
	m framework.NodeToStatusMap) (*framework.PostFilterResult, *framework.Status) {

	nominatedNodeName, err := pl.preempt(ctx, state, pod, m)
	if err != nil {
		return nil, framework.NewStatus(framework.Error, err.Error())
	}
	if nominatedNodeName == "" {
		return nil, framework.NewStatus(framework.Unschedulable)
	}

	return &framework.PostFilterResult{NominatedNodeName: nominatedNodeName}, framework.NewStatus(framework.Success)
}

// preempt finds nodes with pods that can be preempted to make room for "pod" to
// schedule. It chooses one of the nodes and preempts the pods on the node and
// returns
//  1) the node name which is picked up for preemption,
//  2) any possible error.
// preempt does not update its snapshot. It uses the same snapshot used in the
// scheduling cycle. This is to avoid a scenario where preempt finds feasible
// nodes without preempting any pod. When there are many pending pods in the
// scheduling queue a nominated pod will go back to the queue and behind
// other pods with the same priority. The nominated pod prevents other pods from
// using the nominated resources and the nominated pod could take a long time
// before it is retried after many other pending pods.
// INFO: @see https://kubernetes.io/zh/docs/concepts/configuration/pod-priority-preemption/
func (pl *DefaultPreemption) preempt(ctx context.Context, state *framework.CycleState, pod *v1.Pod,
	m framework.NodeToStatusMap) (string, error) {
	nodeLister := pl.fh.SnapshotSharedLister().NodeInfos()

	// 1) Ensure the preemptor is eligible to preempt other pods.
	if !PodEligibleToPreemptOthers(pod, nodeLister, m[pod.Status.NominatedNodeName]) {
		klog.V(5).Infof("Pod %v/%v is not eligible for more preemption.", pod.Namespace, pod.Name)
		return "", nil
	}

	// 2) Find all preemption candidates.

	// 3) Interact with registered Extenders to filter out some candidates if needed.

	// 4) Find the best candidate.

	// 5) Perform preparation work before nominating the selected candidate.

}

// New initializes a new plugin and returns it.
func New(_ runtime.Object, fh framework.FrameworkHandle) (framework.Plugin, error) {
	pl := DefaultPreemption{
		fh:        fh,
		pdbLister: getPDBLister(fh.SharedInformerFactory()),
	}

	return &pl, nil
}

// PodEligibleToPreemptOthers determines whether this pod should be considered
// for preempting other pods or not. If this pod has already preempted other
// pods and those are in their graceful termination period, it shouldn't be
// considered for preemption.
// We look at the node that is nominated for this pod and as long as there are
// terminating pods on the node, we don't consider this for preempting more pods.
// INFO: @see https://kubernetes.io/zh/docs/concepts/configuration/pod-priority-preemption/#non-preempting-priority-class
func PodEligibleToPreemptOthers(pod *v1.Pod, nodeInfos framework.NodeInfoLister, nominatedNodeStatus *framework.Status) bool {
	// INFO: 非抢占式的 pod 不需要抢占，返回 false
	if pod.Spec.PreemptionPolicy != nil && *pod.Spec.PreemptionPolicy == v1.PreemptNever {
		klog.V(5).Infof("Pod %v/%v is not eligible for preemption because it has a preemptionPolicy of %v", pod.Namespace, pod.Name, v1.PreemptNever)
		return false
	}

	nominatedNodeName := pod.Status.NominatedNodeName
	if len(nominatedNodeName) > 0 {
		// If the pod's nominated node is considered as UnschedulableAndUnresolvable by the filters,
		// then the pod should be considered for preempting again.
		// INFO: nominatedNodeName 是不可调度的，可以抢占？
		if nominatedNodeStatus.Code() == framework.UnschedulableAndUnresolvable {
			return true
		}

		// nominatedNodeName 有 terminating pod 了，则不需要抢占
		if nodeInfo, _ := nodeInfos.Get(nominatedNodeName); nodeInfo != nil {
			podPriority := podutil.GetPodPriority(pod)
			for _, p := range nodeInfo.Pods {
				if p.Pod.DeletionTimestamp != nil && podutil.GetPodPriority(p.Pod) < podPriority {
					// There is a terminating pod on the nominated node.
					return false
				}
			}
		}
	}

	return true
}
