package framework

import (
	"k8s-lx1036/k8s/scheduler/pkg/kube-batch/pkg/scheduler/api"
	"k8s-lx1036/k8s/scheduler/pkg/kube-batch/pkg/scheduler/conf"
	"k8s.io/apimachinery/pkg/types"
)

// Session information for the current session
type Session struct {
	UID types.UID

	cache cache.Cache

	Jobs    map[api.JobID]*api.JobInfo
	Nodes   map[string]*api.NodeInfo
	Queues  map[api.QueueID]*api.QueueInfo
	Backlog []*api.JobInfo
	Tiers   []conf.Tier

	plugins          map[string]Plugin
	eventHandlers    []*EventHandler
	jobOrderFns      map[string]api.CompareFn
	queueOrderFns    map[string]api.CompareFn
	taskOrderFns     map[string]api.CompareFn
	predicateFns     map[string]api.PredicateFn
	preemptableFns   map[string]api.EvictableFn
	reclaimableFns   map[string]api.EvictableFn
	overusedFns      map[string]api.ValidateFn
	jobReadyFns      map[string]api.ValidateFn
	jobPipelinedFns  map[string]api.ValidateFn
	jobValidFns      map[string]api.ValidateExFn
	nodePrioritizers map[string][]priorities.PriorityConfig
}
