package defaultpreemption

import (
	"context"

	framework "k8s-lx1036/k8s/scheduler/pkg/scheduler/framework/v1alpha1"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	policylisters "k8s.io/client-go/listers/policy/v1beta1"
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
func (pl *DefaultPreemption) preempt(ctx context.Context, state *framework.CycleState, pod *v1.Pod,
	m framework.NodeToStatusMap) (string, error) {

	// 1) Ensure the preemptor is eligible to preempt other pods.

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
