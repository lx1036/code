package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	componentbaseconfig "k8s.io/component-base/config"
)

const (
	// SchedulerDefaultLockObjectNamespace defines default scheduler lock object namespace ("kube-system")
	SchedulerDefaultLockObjectNamespace string = metav1.NamespaceSystem

	// SchedulerDefaultLockObjectName defines default scheduler lock object name ("kube-scheduler")
	SchedulerDefaultLockObjectName = "kube-scheduler"

	// SchedulerPolicyConfigMapKey defines the key of the element in the
	// scheduler's policy ConfigMap that contains scheduler's policy config.
	SchedulerPolicyConfigMapKey = "policy.cfg"

	// SchedulerDefaultProviderName defines the default provider names
	SchedulerDefaultProviderName = "DefaultProvider"

	// DefaultInsecureSchedulerPort is the default port for the scheduler status server.
	// May be overridden by a flag at startup.
	// Deprecated: use the secure KubeSchedulerPort instead.
	DefaultInsecureSchedulerPort = 10251

	// DefaultKubeSchedulerPort is the default port for the scheduler status server.
	// May be overridden by a flag at startup.
	DefaultKubeSchedulerPort = 10259
)

const (
	// DefaultPercentageOfNodesToScore defines the percentage of nodes of all nodes
	// that once found feasible, the scheduler stops looking for more nodes.
	// A value of 0 means adaptive, meaning the scheduler figures out a proper default.
	DefaultPercentageOfNodesToScore = 0
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// KubeSchedulerConfiguration configures a scheduler
type KubeSchedulerConfiguration struct {
	metav1.TypeMeta

	// ClientConnection specifies the kubeconfig file and client connection
	// settings for the proxy server to use when communicating with the apiserver.
	ClientConnection componentbaseconfig.ClientConnectionConfiguration

	// LeaderElection defines the configuration of leader election client.
	LeaderElection componentbaseconfig.LeaderElectionConfiguration

	// Profiles are scheduling profiles that kube-scheduler supports. Pods can
	// choose to be scheduled under a particular profile by setting its associated
	// scheduler name. Pods that don't specify any scheduler name are scheduled
	// with the "default-scheduler" profile, if present here.
	Profiles []KubeSchedulerProfile

	// PercentageOfNodesToScore is the percentage of all nodes that once found feasible
	// for running a pod, the scheduler stops its search for more feasible nodes in
	// the cluster. This helps improve scheduler's performance. Scheduler always tries to find
	// at least "minFeasibleNodesToFind" feasible nodes no matter what the value of this flag is.
	// Example: if the cluster size is 500 nodes and the value of this flag is 30,
	// then scheduler stops finding further feasible nodes once it finds 150 feasible ones.
	// When the value is 0, default percentage (5%--50% based on the size of the cluster) of the
	// nodes will be scored.
	PercentageOfNodesToScore int32

	// PodInitialBackoffSeconds is the initial backoff for unschedulable pods.
	// If specified, it must be greater than 0. If this value is null, the default value (1s)
	// will be used.
	PodInitialBackoffSeconds int64

	// PodMaxBackoffSeconds is the max backoff for unschedulable pods.
	// If specified, it must be greater than or equal to podInitialBackoffSeconds. If this value is null,
	// the default value (10s) will be used.
	PodMaxBackoffSeconds int64
}

// KubeSchedulerProfile is a scheduling profile.
type KubeSchedulerProfile struct {
	SchedulerName string

	Plugins *Plugins

	PluginConfig []PluginConfig
}

// PluginConfig specifies arguments that should be passed to a plugin at the time of initialization.
// A plugin that is invoked at multiple extension points is initialized once. Args can have arbitrary structure.
// It is up to the plugin to process these Args.
type PluginConfig struct {
	// Name defines the name of plugin being configured
	Name string
	// Args defines the arguments passed to the plugins at the time of initialization. Args can have arbitrary structure.
	Args runtime.Object
}

// Plugins include multiple extension points. When specified, the list of plugins for
// a particular extension point are the only ones enabled. If an extension point is
// omitted from the config, then the default set of plugins is used for that extension point.
// Enabled plugins are called in the order specified here, after default plugins. If they need to
// be invoked before default plugins, default plugins must be disabled and re-enabled here in desired order.
type Plugins struct {
	// QueueSort is a list of plugins that should be invoked when sorting pods in the scheduling queue.
	QueueSort PluginSet

	// PreFilter is a list of plugins that should be invoked at "PreFilter" extension point of the scheduling framework.
	PreFilter PluginSet

	// Filter is a list of plugins that should be invoked when filtering out nodes that cannot run the Pod.
	Filter PluginSet

	// PostFilter is a list of plugins that are invoked after filtering phase, no matter whether filtering succeeds or not.
	PostFilter PluginSet

	// PreScore is a list of plugins that are invoked before scoring.
	PreScore PluginSet

	// Score is a list of plugins that should be invoked when ranking nodes that have passed the filtering phase.
	Score PluginSet

	// Reserve is a list of plugins invoked when reserving/unreserving resources
	// after a node is assigned to run the pod.
	Reserve PluginSet

	// Permit is a list of plugins that control binding of a Pod. These plugins can prevent or delay binding of a Pod.
	Permit PluginSet

	// PreBind is a list of plugins that should be invoked before a pod is bound.
	PreBind PluginSet

	// Bind is a list of plugins that should be invoked at "Bind" extension point of the scheduling framework.
	// The scheduler call these plugins in order. Scheduler skips the rest of these plugins as soon as one returns success.
	Bind PluginSet

	// PostBind is a list of plugins that should be invoked after a pod is successfully bound.
	PostBind PluginSet
}

// PluginSet specifies enabled and disabled plugins for an extension point.
// If an array is empty, missing, or nil, default plugins at that extension point will be used.
type PluginSet struct {
	// Enabled specifies plugins that should be enabled in addition to default plugins.
	// These are called after default plugins and in the same order specified here.
	Enabled []Plugin
	// Disabled specifies default plugins that should be disabled.
	// When all default plugins need to be disabled, an array containing only one "*" should be provided.
	Disabled []Plugin
}

// Plugin specifies a plugin name and its weight when applicable. Weight is used only for Score plugins.
type Plugin struct {
	// Name defines the name of plugin
	Name string
	// Weight defines the weight of plugin, only used for Score plugins.
	Weight int32
}

func appendPluginSet(dst *PluginSet, src *PluginSet) *PluginSet {
	if dst == nil {
		dst = &PluginSet{}
	}
	if src != nil {
		dst.Enabled = append(dst.Enabled, src.Enabled...)
		dst.Disabled = append(dst.Disabled, src.Disabled...)
	}
	return dst
}

// SchedulerAlgorithmSource is the source of a scheduler algorithm. One source
// field must be specified, and source fields are mutually exclusive.
type SchedulerAlgorithmSource struct {
	// Policy is a policy based algorithm source.
	Policy *SchedulerPolicySource
	// Provider is the name of a scheduling algorithm provider to use.
	Provider *string
}

// SchedulerPolicySource configures a means to obtain a scheduler Policy. One
// source field must be specified, and source fields are mutually exclusive.
type SchedulerPolicySource struct {
	// File is a file policy source.
	File *SchedulerPolicyFileSource
	// ConfigMap is a config map policy source.
	ConfigMap *SchedulerPolicyConfigMapSource
}

// SchedulerPolicyFileSource is a policy serialized to disk and accessed via
// path.
type SchedulerPolicyFileSource struct {
	// Path is the location of a serialized policy.
	Path string
}

// SchedulerPolicyConfigMapSource is a policy serialized into a config map value
// under the SchedulerPolicyConfigMapKey key.
type SchedulerPolicyConfigMapSource struct {
	// Namespace is the namespace of the policy config map.
	Namespace string
	// Name is the name of the policy config map.
	Name string
}

// Extender holds the parameters used to communicate with the extender. If a verb is unspecified/empty,
// it is assumed that the extender chose not to provide that extension.
type Extender struct {
	// URLPrefix at which the extender is available
	URLPrefix string
	// Verb for the filter call, empty if not supported. This verb is appended to the URLPrefix when issuing the filter call to extender.
	FilterVerb string
	// Verb for the preempt call, empty if not supported. This verb is appended to the URLPrefix when issuing the preempt call to extender.
	PreemptVerb string
	// Verb for the prioritize call, empty if not supported. This verb is appended to the URLPrefix when issuing the prioritize call to extender.
	PrioritizeVerb string
	// The numeric multiplier for the node scores that the prioritize call generates.
	// The weight should be a positive integer
	Weight int64
	// Verb for the bind call, empty if not supported. This verb is appended to the URLPrefix when issuing the bind call to extender.
	// If this method is implemented by the extender, it is the extender's responsibility to bind the pod to apiserver. Only one extender
	// can implement this function.
	BindVerb string
	// EnableHTTPS specifies whether https should be used to communicate with the extender
	EnableHTTPS bool
	// TLSConfig specifies the transport layer security config
	TLSConfig *ExtenderTLSConfig
	// HTTPTimeout specifies the timeout duration for a call to the extender. Filter timeout fails the scheduling of the pod. Prioritize
	// timeout is ignored, k8s/other extenders priorities are used to select the node.
	HTTPTimeout metav1.Duration
	// NodeCacheCapable specifies that the extender is capable of caching node information,
	// so the scheduler should only send minimal information about the eligible nodes
	// assuming that the extender already cached full details of all nodes in the cluster
	NodeCacheCapable bool
	// ManagedResources is a list of extended resources that are managed by
	// this extender.
	// - A pod will be sent to the extender on the Filter, Prioritize and Bind
	//   (if the extender is the binder) phases iff the pod requests at least
	//   one of the extended resources in this list. If empty or unspecified,
	//   all pods will be sent to this extender.
	// - If IgnoredByScheduler is set to true for a resource, kube-scheduler
	//   will skip checking the resource in predicates.
	// +optional
	ManagedResources []ExtenderManagedResource
	// Ignorable specifies if the extender is ignorable, i.e. scheduling should not
	// fail when the extender returns an error or is not reachable.
	Ignorable bool
}

// ExtenderManagedResource describes the arguments of extended resources
// managed by an extender.
type ExtenderManagedResource struct {
	// Name is the extended resource name.
	Name string
	// IgnoredByScheduler indicates whether kube-scheduler should ignore this
	// resource when applying predicates.
	IgnoredByScheduler bool
}

// ExtenderTLSConfig contains settings to enable TLS with extender
type ExtenderTLSConfig struct {
	// Server should be accessed without verifying the TLS certificate. For testing only.
	Insecure bool
	// ServerName is passed to the server for SNI and is used in the client to check server
	// certificates against. If ServerName is empty, the hostname used to contact the
	// server is used.
	ServerName string

	// Server requires TLS client certificate authentication
	CertFile string
	// Server requires TLS client certificate authentication
	KeyFile string
	// Trusted root certificates for server
	CAFile string

	// CertData holds PEM-encoded bytes (typically read from a client certificate file).
	// CertData takes precedence over CertFile
	CertData []byte
	// KeyData holds PEM-encoded bytes (typically read from a client certificate key file).
	// KeyData takes precedence over KeyFile
	KeyData []byte
	// CAData holds PEM-encoded bytes (typically read from a root certificates bundle).
	// CAData takes precedence over CAFile
	CAData []byte
}
