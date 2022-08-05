package testing

import (
	"k8s-lx1036/k8s/scheduler/pkg/apis/config/scheme"
	configv1 "k8s-lx1036/k8s/scheduler/pkg/apis/config/v1"
	"k8s-lx1036/k8s/scheduler/pkg/framework/runtime"
)

// RegisterPluginFunc is a function signature used in method RegisterFilterPlugin()
// to register a Filter Plugin to a given registry.
type RegisterPluginFunc func(reg *runtime.Registry, profile *configv1.KubeSchedulerProfile)

func RegisterFilterPlugin(pluginName string, pluginNewFunc runtime.PluginFactory) RegisterPluginFunc {
	return RegisterPluginAsExtensions(pluginName, pluginNewFunc, "Filter")
}

func RegisterQueueSortPlugin(pluginName string, pluginNewFunc runtime.PluginFactory) RegisterPluginFunc {
	return RegisterPluginAsExtensions(pluginName, pluginNewFunc, "QueueSort")
}

func RegisterBindPlugin(pluginName string, pluginNewFunc runtime.PluginFactory) RegisterPluginFunc {
	return RegisterPluginAsExtensions(pluginName, pluginNewFunc, "Bind")
}

func RegisterPluginAsExtensions(pluginName string, pluginNewFunc runtime.PluginFactory, extensions ...string) RegisterPluginFunc {
	return RegisterPluginAsExtensionsWithWeight(pluginName, 1, pluginNewFunc, extensions...)
}

var configDecoder = scheme.Codecs.UniversalDecoder()

func RegisterPluginAsExtensionsWithWeight(pluginName string, weight int32, pluginNewFunc runtime.PluginFactory,
	extensions ...string) RegisterPluginFunc {
	return func(r *runtime.Registry, profile *configv1.KubeSchedulerProfile) {
		r.Register(pluginName, pluginNewFunc)
		for _, extension := range extensions { // 修改具体的 profile.Plugins Enabled
			pluginSet := getPluginSetByExtension(profile.Plugins, extension)
			if pluginSet == nil {
				continue
			}
			pluginSet.Enabled = append(pluginSet.Enabled, configv1.Plugin{Name: pluginName, Weight: weight})
		}

		// INFO: 这里还得琢磨下???
		gvk := configv1.SchemeGroupVersion.WithKind(pluginName + "Args") // Kind = pluginName+"Args" ???
		if args, _, err := configDecoder.Decode(nil, &gvk, nil); err == nil {
			profile.PluginConfig = append(profile.PluginConfig, configv1.PluginConfig{
				Name: pluginName,
				Args: args,
			})
		}
	}
}

func getPluginSetByExtension(plugins *configv1.Plugins, extension string) *configv1.PluginSet {
	switch extension {
	case "QueueSort":
		return &plugins.QueueSort
	case "Filter":
		return &plugins.Filter
	case "PreFilter":
		return &plugins.PreFilter
	case "PreScore":
		return &plugins.PreScore
	case "Score":
		return &plugins.Score
	case "Bind":
		return &plugins.Bind
	case "Reserve":
		return &plugins.Reserve
	case "Permit":
		return &plugins.Permit
	case "PreBind":
		return &plugins.PreBind
	case "PostBind":
		return &plugins.PostBind
	default:
		return nil
	}
}

func NewFramework(pluginFunc []RegisterPluginFunc, profileName string, opts ...runtime.Option) (*runtime.Framework, error) {
	registry := runtime.Registry{}
	profile := &configv1.KubeSchedulerProfile{
		SchedulerName: profileName,
		Plugins:       &configv1.Plugins{},
	}
	for _, f := range pluginFunc {
		f(&registry, profile)
	}

	return runtime.NewFramework(registry, profile, opts...)
}
