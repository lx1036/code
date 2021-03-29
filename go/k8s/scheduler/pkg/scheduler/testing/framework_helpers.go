package testing

import (
	"k8s-lx1036/k8s/scheduler/pkg/scheduler/framework/runtime"
)

// RegisterPluginFunc is a function signature used in method RegisterFilterPlugin()
// to register a Filter Plugin to a given registry.
type RegisterPluginFunc func(reg *runtime.Registry, plugins *schedulerapi.Plugins, pluginConfigs []schedulerapi.PluginConfig)
