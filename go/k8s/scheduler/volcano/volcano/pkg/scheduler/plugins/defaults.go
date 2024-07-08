package plugins

import (
	"k8s-lx1036/k8s/scheduler/volcano/volcano/pkg/scheduler/conf"
)

func ApplyPluginConfDefaults(option *conf.PluginOption) {
	t := true

	if option.EnabledJobOrder == nil {
		option.EnabledJobOrder = &t
	}
	if option.EnabledJobReady == nil {
		option.EnabledJobReady = &t
	}
	if option.EnabledJobPipelined == nil {
		option.EnabledJobPipelined = &t
	}
	if option.EnabledJobEnqueued == nil {
		option.EnabledJobEnqueued = &t
	}
	if option.EnabledTaskOrder == nil {
		option.EnabledTaskOrder = &t
	}
	if option.EnabledPreemptable == nil {
		option.EnabledPreemptable = &t
	}
	if option.EnabledReclaimable == nil {
		option.EnabledReclaimable = &t
	}
	if option.EnablePreemptive == nil {
		option.EnablePreemptive = &t
	}
	if option.EnabledQueueOrder == nil {
		option.EnabledQueueOrder = &t
	}
	if option.EnabledPredicate == nil {
		option.EnabledPredicate = &t
	}
	if option.EnabledBestNode == nil {
		option.EnabledBestNode = &t
	}
	if option.EnabledNodeOrder == nil {
		option.EnabledNodeOrder = &t
	}
	if option.EnabledTargetJob == nil {
		option.EnabledTargetJob = &t
	}
	if option.EnabledReservedNodes == nil {
		option.EnabledReservedNodes = &t
	}
	if option.EnabledVictim == nil {
		option.EnabledVictim = &t
	}
	if option.EnabledJobStarving == nil {
		option.EnabledJobStarving = &t
	}
	if option.EnabledOverused == nil {
		option.EnabledOverused = &t
	}
	if option.EnabledAllocatable == nil {
		option.EnabledAllocatable = &t
	}
}
