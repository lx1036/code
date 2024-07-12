package framework

import (
	"time"

	"k8s-lx1036/k8s/scheduler/volcano/volcano/pkg/scheduler/api"
	"k8s-lx1036/k8s/scheduler/volcano/volcano/pkg/scheduler/cache"
	"k8s-lx1036/k8s/scheduler/volcano/volcano/pkg/scheduler/conf"
	"k8s-lx1036/k8s/scheduler/volcano/volcano/pkg/scheduler/metrics"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	k8sframework "k8s.io/kubernetes/pkg/scheduler/framework"
	"volcano.sh/apis/pkg/apis/scheduling"
)

type Session struct {
	UID types.UID

	Tiers          []conf.Tier
	plugins        map[string]Plugin
	Configurations []conf.Configuration

	Queues         map[api.QueueID]*api.QueueInfo
	Jobs           map[api.JobID]*api.JobInfo
	Nodes          map[string]*api.NodeInfo
	NodeList       []*api.NodeInfo
	CSINodesStatus map[string]*api.CSINodeStatusInfo
	RevocableNodes map[string]*api.NodeInfo
	NamespaceInfo  map[api.NamespaceName]*api.NamespaceInfo

	// NodeMap is like Nodes except that it uses k8s NodeInfo api and should only
	// be used in k8s compatable api scenarios such as in predicates and nodeorder plugins.
	NodeMap   map[string]*k8sframework.NodeInfo
	PodLister *PodLister

	TotalResource *api.Resource
	// podGroupStatus cache podgroup status during schedule
	// This should not be mutated after initiated
	podGroupStatus map[api.JobID]scheduling.PodGroupStatus

	taskOrderFns   map[string]api.CompareFn
	jobOrderFns    map[string]api.CompareFn
	preemptableFns map[string]api.EvictableFn
	jobStarvingFns map[string]api.ValidateFn
	jobValidFns    map[string]api.ValidateExFn
}

func OpenSession(cache cache.Cache, tiers []conf.Tier, configurations []conf.Configuration) *Session {
	ssn := openSession(cache)
	ssn.Tiers = tiers
	ssn.Configurations = configurations
	ssn.NodeMap = GenerateNodeMapAndSlice(ssn.Nodes)
	ssn.PodLister = NewPodLister(ssn)

	for _, tier := range tiers {
		for _, plugin := range tier.Plugins {
			if pb, found := GetPluginBuilder(plugin.Name); !found {
				klog.Errorf("Failed to get plugin %s.", plugin.Name)
			} else {
				p := pb(plugin.Arguments)
				ssn.plugins[p.Name()] = p
				onSessionOpenStart := time.Now()
				p.OnSessionOpen(ssn)
				metrics.UpdatePluginDuration(p.Name(), metrics.OnSessionOpen, metrics.Duration(onSessionOpenStart))
			}
		}
	}

	return ssn
}

func openSession(cache cache.Cache) *Session {
	ssn := &Session{
		Jobs:  map[api.JobID]*api.JobInfo{},
		Nodes: map[string]*api.NodeInfo{},
	}

	snapshot := cache.Snapshot()
	ssn.Jobs = snapshot.Jobs
	for _, job := range ssn.Jobs {
		if job.PodGroup != nil {
			ssn.podGroupStatus[job.UID] = *job.PodGroup.Status.DeepCopy()
		}

		// job not valid
		if vjr := ssn.JobValid(job); vjr != nil {
			jc := &scheduling.PodGroupCondition{
				Type:               scheduling.PodGroupUnschedulableType,
				Status:             v1.ConditionTrue,
				LastTransitionTime: metav1.Now(),
				TransitionID:       string(ssn.UID),
				Reason:             vjr.Reason,
				Message:            vjr.Message,
			}

			if err := ssn.UpdatePodGroupCondition(job, jc); err != nil {
				klog.Errorf("Failed to update job condition: %v", err)
			}

			delete(ssn.Jobs, job.UID)
		}
	}

	ssn.NodeList = util.GetNodeList(snapshot.Nodes, snapshot.NodeList)
	ssn.Nodes = snapshot.Nodes
	ssn.CSINodesStatus = snapshot.CSINodesStatus
	ssn.RevocableNodes = snapshot.RevocableNodes
	ssn.Queues = snapshot.Queues
	ssn.NamespaceInfo = snapshot.NamespaceInfo
	// calculate all nodes' resource only once in each schedule cycle, other plugins can clone it when need
	for _, n := range ssn.Nodes {
		ssn.TotalResource.Add(n.Allocatable)
	}

	klog.V(3).Infof("Open Session %v with <%d> Job and <%d> Queues",
		ssn.UID, len(ssn.Jobs), len(ssn.Queues))

	return ssn
}

func CloseSession(ssn *Session) {
	for _, plugin := range ssn.plugins {
		onSessionCloseStart := time.Now()
		plugin.OnSessionClose(ssn)
		metrics.UpdatePluginDuration(plugin.Name(), metrics.OnSessionClose, metrics.Duration(onSessionCloseStart))
	}

	closeSession(ssn)
}

func closeSession(ssn *Session) {
	// TODO

	ssn.Jobs = nil
	ssn.Nodes = nil
	//ssn.RevocableNodes = nil
	ssn.plugins = nil
	//ssn.eventHandlers = nil
	//ssn.jobOrderFns = nil
	//ssn.queueOrderFns = nil
	//ssn.clusterOrderFns = nil
	//ssn.NodeList = nil
	//ssn.TotalResource = nil

	klog.V(3).Infof("Close Session %v", ssn.UID)
}
