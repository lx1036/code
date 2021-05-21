package algorithmprovider

import (
	"k8s-lx1036/k8s/scheduler/pkg/scheduler/apis/config"

	"k8s-lx1036/k8s/scheduler/pkg/scheduler/framework/plugins/defaultbinder"
	"k8s-lx1036/k8s/scheduler/pkg/scheduler/framework/plugins/defaultpreemption"
	"k8s-lx1036/k8s/scheduler/pkg/scheduler/framework/plugins/imagelocality"
	"k8s-lx1036/k8s/scheduler/pkg/scheduler/framework/plugins/interpodaffinity"
	"k8s-lx1036/k8s/scheduler/pkg/scheduler/framework/plugins/nodeaffinity"
	"k8s-lx1036/k8s/scheduler/pkg/scheduler/framework/plugins/nodename"
	"k8s-lx1036/k8s/scheduler/pkg/scheduler/framework/plugins/nodeports"
	"k8s-lx1036/k8s/scheduler/pkg/scheduler/framework/plugins/nodepreferavoidpods"
	"k8s-lx1036/k8s/scheduler/pkg/scheduler/framework/plugins/noderesources"
	"k8s-lx1036/k8s/scheduler/pkg/scheduler/framework/plugins/nodeunschedulable"
	"k8s-lx1036/k8s/scheduler/pkg/scheduler/framework/plugins/nodevolumelimits"
	"k8s-lx1036/k8s/scheduler/pkg/scheduler/framework/plugins/podtopologyspread"
	"k8s-lx1036/k8s/scheduler/pkg/scheduler/framework/plugins/queuesort"
	"k8s-lx1036/k8s/scheduler/pkg/scheduler/framework/plugins/tainttoleration"
	"k8s-lx1036/k8s/scheduler/pkg/scheduler/framework/plugins/volumebinding"
	"k8s-lx1036/k8s/scheduler/pkg/scheduler/framework/plugins/volumerestrictions"
	"k8s-lx1036/k8s/scheduler/pkg/scheduler/framework/plugins/volumezone"
)

// ClusterAutoscalerProvider defines the default autoscaler provider
const ClusterAutoscalerProvider = "ClusterAutoscalerProvider"

// Registry is a collection of all available algorithm providers.
type Registry map[string]*config.Plugins

// NewRegistry returns an algorithm provider registry instance.
func NewRegistry() Registry {
	defaultConfig := getDefaultPlugins()
	//applyFeatureGates(defaultConfig)

	//caConfig := getClusterAutoscalerConfig()
	//applyFeatureGates(caConfig)

	return Registry{
		config.SchedulerDefaultProviderName: defaultConfig,
		//ClusterAutoscalerProvider:                 caConfig,
	}
}

func getDefaultPlugins() *config.Plugins {
	return &config.Plugins{
		QueueSort: &config.PluginSet{
			Enabled: []config.Plugin{
				{Name: queuesort.Name},
			},
		},
		PreFilter: &config.PluginSet{
			Enabled: []config.Plugin{
				{Name: noderesources.FitName},
				{Name: nodeports.Name},
				{Name: podtopologyspread.Name},
				{Name: interpodaffinity.Name},
				{Name: volumebinding.Name},
			},
		},
		Filter: &config.PluginSet{
			Enabled: []config.Plugin{
				{Name: nodeunschedulable.Name},
				{Name: noderesources.FitName},
				{Name: nodename.Name},
				{Name: nodeports.Name},
				{Name: nodeaffinity.Name},
				{Name: volumerestrictions.Name},
				{Name: tainttoleration.Name},
				{Name: nodevolumelimits.EBSName},
				{Name: nodevolumelimits.GCEPDName},
				{Name: nodevolumelimits.CSIName},
				{Name: nodevolumelimits.AzureDiskName},
				{Name: volumebinding.Name},
				{Name: volumezone.Name},
				{Name: podtopologyspread.Name},
				{Name: interpodaffinity.Name},
			},
		},
		PostFilter: &config.PluginSet{
			Enabled: []config.Plugin{
				{Name: defaultpreemption.Name},
			},
		},
		PreScore: &config.PluginSet{
			Enabled: []config.Plugin{
				{Name: interpodaffinity.Name},
				{Name: podtopologyspread.Name},
				{Name: tainttoleration.Name},
			},
		},
		Score: &config.PluginSet{
			Enabled: []config.Plugin{
				{Name: noderesources.BalancedAllocationName, Weight: 1},
				{Name: imagelocality.Name, Weight: 1},
				{Name: interpodaffinity.Name, Weight: 1},
				{Name: noderesources.LeastAllocatedName, Weight: 1},
				{Name: nodeaffinity.Name, Weight: 1},
				{Name: nodepreferavoidpods.Name, Weight: 10000},
				// Weight is doubled because:
				// - This is a score coming from user preference.
				// - It makes its signal comparable to NodeResourcesLeastAllocated.
				{Name: podtopologyspread.Name, Weight: 2},
				{Name: tainttoleration.Name, Weight: 1},
			},
		},
		Reserve: &config.PluginSet{
			Enabled: []config.Plugin{
				{Name: volumebinding.Name},
			},
		},
		PreBind: &config.PluginSet{
			Enabled: []config.Plugin{
				{Name: volumebinding.Name},
			},
		},
		Bind: &config.PluginSet{
			Enabled: []config.Plugin{
				{Name: defaultbinder.Name},
			},
		},
	}
}
