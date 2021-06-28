package conf

/*
`
actions: "allocate, backfill"
tiers:
- plugins:
  - name: priority
  - name: gang
- plugins:
  - name: drf
  - name: predicates
  - name: proportion
  - name: nodeorder
`
*/

// SchedulerConfiguration defines the configuration of scheduler.
type SchedulerConfiguration struct {
	// Actions defines the actions list of scheduler in order
	Actions string `yaml:"actions"`
	// Tiers defines plugins in different tiers
	Tiers []Tier `yaml:"tiers"`
}

// Tier defines plugin tier
type Tier struct {
	Plugins []PluginOption `yaml:"plugins"`
}

// PluginOption defines the options of plugin
type PluginOption struct {
	// The name of Plugin
	Name string `yaml:"name"`
	// EnabledJobOrder defines whether jobOrderFn is enabled
	EnabledJobOrder *bool `yaml:"enableJobOrder"`
	// EnabledJobReady defines whether jobReadyFn is enabled
	EnabledJobReady *bool `yaml:"enableJobReady"`
	// EnabledJobPipelined defines whether jobPipelinedFn is enabled
	EnabledJobPipelined *bool `yaml:"enableJobPipelined"`
	// EnabledTaskOrder defines whether taskOrderFn is enabled
	EnabledTaskOrder *bool `yaml:"enableTaskOrder"`
	// EnabledPreemptable defines whether preemptableFn is enabled
	EnabledPreemptable *bool `yaml:"enablePreemptable"`
	// EnabledReclaimable defines whether reclaimableFn is enabled
	EnabledReclaimable *bool `yaml:"enableReclaimable"`
	// EnabledQueueOrder defines whether queueOrderFn is enabled
	EnabledQueueOrder *bool `yaml:"enableQueueOrder"`
	// EnabledPredicate defines whether predicateFn is enabled
	EnabledPredicate *bool `yaml:"enablePredicate"`
	// EnabledNodeOrder defines whether NodeOrderFn is enabled
	EnabledNodeOrder *bool `yaml:"enableNodeOrder"`
	// Arguments defines the different arguments that can be given to different plugins
	Arguments map[string]string `yaml:"arguments"`
}

// Configuration is configuration of action
type Configuration struct {
	// Name is name of action
	Name string `yaml:"name"`
	// Arguments defines the different arguments that can be given to specified action
	Arguments map[string]string `yaml:"arguments"`
}
