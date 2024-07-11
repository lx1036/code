package framework

import (
	"k8s-lx1036/k8s/scheduler/volcano/volcano/pkg/scheduler/api"
	"k8s-lx1036/k8s/scheduler/volcano/volcano/pkg/scheduler/cache"
	"k8s-lx1036/k8s/scheduler/volcano/volcano/pkg/scheduler/conf"
	"k8s-lx1036/k8s/scheduler/volcano/volcano/pkg/scheduler/metrics"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	k8sframework "k8s.io/kubernetes/pkg/scheduler/framework"
	"time"
)

type Session struct {
	UID types.UID

	Tiers          []conf.Tier
	plugins        map[string]Plugin
	Configurations []conf.Configuration

	Queues map[api.QueueID]*api.QueueInfo
	Jobs   map[api.JobID]*api.JobInfo
	Nodes  map[string]*api.NodeInfo

	// NodeMap is like Nodes except that it uses k8s NodeInfo api and should only
	// be used in k8s compatable api scenarios such as in predicates and nodeorder plugins.
	NodeMap   map[string]*k8sframework.NodeInfo
	PodLister *PodLister

	taskOrderFns   map[string]api.CompareFn
	jobOrderFns    map[string]api.CompareFn
	preemptableFns map[string]api.EvictableFn
	jobStarvingFns map[string]api.ValidateFn
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
	ssn := &Session{}

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
