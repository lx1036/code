package plugins

import (
	"k8s-lx1036/k8s/scheduler/volcano/volcano/pkg/scheduler/framework"
	"k8s-lx1036/k8s/scheduler/volcano/volcano/pkg/scheduler/plugins/priority"
)

func init() {
	// Plugins for Jobs
	framework.RegisterPluginBuilder(priority.PluginName, priority.New)

	// Plugins for Queues

	// Plugins for Extender

	// Plugins for ResourceQuota
}
