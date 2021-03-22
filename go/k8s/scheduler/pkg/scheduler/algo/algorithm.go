package algo

import (
	"context"
	"time"

	"k8s-lx1036/k8s/scheduler/pkg/scheduler/framework/v1alpha1"
	internalcache "k8s-lx1036/k8s/scheduler/pkg/scheduler/internal/cache"
	"k8s-lx1036/k8s/scheduler/pkg/scheduler/profile"

	v1 "k8s.io/api/core/v1"
	corelisters "k8s.io/client-go/listers/core/v1"
	utiltrace "k8s.io/utils/trace"
)

// ScheduleAlgorithm is an interface implemented by things that know how to schedule pods
// onto machines.
// TODO: Rename this type.
type ScheduleAlgorithm interface {
	Schedule(context.Context, *profile.Profile, *v1alpha1.CycleState, *v1.Pod) (scheduleResult ScheduleResult, err error)
	// Extenders returns a slice of extender config. This is exposed for
	// testing.
	Extenders() []v1alpha1.Extender
}

// ScheduleResult represents the result of one pod scheduled. It will contain
// the final selected Node, along with the selected intermediate information.
type ScheduleResult struct {
	// Name of the scheduler suggest host
	SuggestedHost string
	// Number of nodes scheduler evaluated on one pod scheduled
	EvaluatedNodes int
	// Number of feasible nodes on one pod scheduled
	FeasibleNodes int
}

type genericScheduler struct {
	cache internalcache.Cache
	//extenders                []v1alpha1.Extender
	nodeInfoSnapshot         *internalcache.Snapshot
	pvcLister                corelisters.PersistentVolumeClaimLister
	disablePreemption        bool
	percentageOfNodesToScore int32
	nextStartNodeIndex       int
}

// Filters the nodes to find the ones that fit the pod based on the framework
// filter plugins and filter extenders.
func (g *genericScheduler) findNodesThatFitPod(ctx context.Context, prof *profile.Profile, state *framework.CycleState, pod *v1.Pod) ([]*v1.Node, framework.NodeToStatusMap, error) {
	filteredNodesStatuses := make(framework.NodeToStatusMap)

	// Run "prefilter" plugins.
	s := prof.RunPreFilterPlugins(ctx, state, pod)

}

// Schedule tries to schedule the given pod to one of the nodes in the node list.
// If it succeeds, it will return the name of the node.
// If it fails, it will return a FitError error with reasons.
func (g *genericScheduler) Schedule(ctx context.Context, prof *profile.Profile, state *v1alpha1.CycleState, pod *v1.Pod) (result ScheduleResult, err error) {
	trace := utiltrace.New("Scheduling", utiltrace.Field{Key: "namespace", Value: pod.Namespace}, utiltrace.Field{Key: "name", Value: pod.Name})
	defer trace.LogIfLong(100 * time.Millisecond)

	// 1. PreFilter plugins
	startPredicateEvalTime := time.Now()
	feasibleNodes, filteredNodesStatuses, err := g.findNodesThatFitPod(ctx, prof, state, pod)
	if err != nil {
		return result, err
	}
	trace.Step("Computing predicates done")

	// 2. Score plugins
	// 2.
	priorityList, err := g.prioritizeNodes(ctx, prof, state, pod, feasibleNodes)
	if err != nil {
		return result, err
	}

	// 3.
	host, err := g.selectHost(priorityList)
	trace.Step("Prioritizing done")

	return ScheduleResult{
		SuggestedHost:  host,
		EvaluatedNodes: len(feasibleNodes) + len(filteredNodesStatuses),
		FeasibleNodes:  len(feasibleNodes),
	}, err
}

func (g *genericScheduler) Extenders() []v1alpha1.Extender {
	return g.extenders
}

// Filters the nodes to find the ones that fit the pod based on the framework
// filter plugins and filter extenders.
// (1) 主要运行 PreFilter/Filter plugin
func (g *genericScheduler) findNodesThatFitPod(ctx context.Context, prof *profile.Profile, state *framework.CycleState, pod *v1.Pod) ([]*v1.Node, framework.NodeToStatusMap, error) {
	filteredNodesStatuses := make(framework.NodeToStatusMap)

	// Run "prefilter" plugins.
	s := prof.RunPreFilterPlugins(ctx, state, pod)

	feasibleNodes, err := g.findNodesThatPassFilters(ctx, prof, state, pod, filteredNodesStatuses)
	if err != nil {
		return nil, nil, err
	}

}

// findNodesThatPassFilters finds the nodes that fit the filter plugins.
// (1) 运行 Filter plugin
func (g *genericScheduler) findNodesThatPassFilters(ctx context.Context, prof *profile.Profile, state *framework.CycleState, pod *v1.Pod, statuses framework.NodeToStatusMap) ([]*v1.Node, error) {

}

// NewGenericScheduler creates a genericScheduler object.
func NewGenericScheduler(
	cache internalcache.Cache,
	nodeInfoSnapshot *internalcache.Snapshot,
	//extenders []v1alpha1.Extender,
	pvcLister corelisters.PersistentVolumeClaimLister,
	disablePreemption bool,
	percentageOfNodesToScore int32) ScheduleAlgorithm {
	return &genericScheduler{
		cache: cache,
		//extenders:                extenders,
		nodeInfoSnapshot:         nodeInfoSnapshot,
		pvcLister:                pvcLister,
		disablePreemption:        disablePreemption,
		percentageOfNodesToScore: percentageOfNodesToScore,
	}
}
