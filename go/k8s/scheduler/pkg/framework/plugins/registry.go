package plugins

import (
	"k8s-lx1036/k8s/scheduler/pkg/framework/plugins/defaultbinder"
	"k8s-lx1036/k8s/scheduler/pkg/framework/plugins/defaultpreemption"
	"k8s-lx1036/k8s/scheduler/pkg/framework/plugins/nodeaffinity"
	"k8s-lx1036/k8s/scheduler/pkg/framework/plugins/nodename"
	"k8s-lx1036/k8s/scheduler/pkg/framework/plugins/nodeports"
	"k8s-lx1036/k8s/scheduler/pkg/framework/plugins/noderesources"
	"k8s-lx1036/k8s/scheduler/pkg/framework/plugins/nodeunschedulable"
	"k8s-lx1036/k8s/scheduler/pkg/framework/plugins/nodevolumelimits"
	"k8s-lx1036/k8s/scheduler/pkg/framework/plugins/queuesort"
	"k8s-lx1036/k8s/scheduler/pkg/framework/runtime"

	"k8s.io/kubernetes/pkg/scheduler/framework/plugins/interpodaffinity"
	"k8s.io/kubernetes/pkg/scheduler/framework/plugins/podtopologyspread"
	"k8s.io/kubernetes/pkg/scheduler/framework/plugins/selectorspread"
	"k8s.io/kubernetes/pkg/scheduler/framework/plugins/tainttoleration"
	"k8s.io/kubernetes/pkg/scheduler/framework/plugins/volumebinding"
	"k8s.io/kubernetes/pkg/scheduler/framework/plugins/volumerestrictions"
)

func NewInTreeRegistry() runtime.Registry {
	return runtime.Registry{
		queuesort.Name: queuesort.New,

		defaultbinder.Name:     defaultbinder.New,
		defaultpreemption.Name: defaultpreemption.New,

		selectorspread.Name: selectorspread.New,
		//imagelocality.Name:                   imagelocality.New,
		tainttoleration.Name:                 tainttoleration.New,
		nodename.Name:                        nodename.New,
		nodeports.Name:                       nodeports.New,
		nodeaffinity.Name:                    nodeaffinity.New,
		podtopologyspread.Name:               podtopologyspread.New,
		nodeunschedulable.Name:               nodeunschedulable.New,
		noderesources.Name:                   noderesources.NewFit,
		noderesources.BalancedAllocationName: noderesources.NewBalancedAllocation,
		volumebinding.Name:                   volumebinding.New,
		volumerestrictions.Name:              volumerestrictions.New,
		//volumezone.Name:                      volumezone.New,
		nodevolumelimits.CSIName: nodevolumelimits.NewCSI,
		interpodaffinity.Name:    interpodaffinity.New,
	}
}
