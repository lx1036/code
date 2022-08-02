package runtime

import (
	"context"
	"fmt"
	"time"

	"k8s-lx1036/k8s/scheduler/pkg/apis/config"
	"k8s-lx1036/k8s/scheduler/pkg/apis/config/scheme"
	schedulerapiv1 "k8s-lx1036/k8s/scheduler/pkg/apis/config/v1"
	"k8s-lx1036/k8s/scheduler/pkg/framework"
	"k8s-lx1036/k8s/scheduler/pkg/framework/parallelize"
	"k8s-lx1036/k8s/scheduler/pkg/metrics"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/informers"
	clientset "k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/events"
)

const (
	postFilter = "PostFilter"
)

type frameworkOptions struct {
	kubeConfig           *restclient.Config
	clientSet            clientset.Interface
	eventRecorder        events.EventRecorder
	informerFactory      informers.SharedInformerFactory
	snapshotSharedLister framework.SharedLister
	metricsRecorder      *metricsRecorder
	profileName          string
	podNominator         framework.PodNominator
	runAllFilters        bool
	captureProfile       CaptureProfile
	clusterEventMap      map[framework.ClusterEvent]sets.String
	parallelizer         parallelize.Parallelizer
}

// Option for the Framework.
type Option func(*frameworkOptions)

// WithPodNominator sets podNominator for the scheduling Framework.
func WithPodNominator(nominator framework.PodNominator) Option {
	return func(o *frameworkOptions) {
		o.podNominator = nominator
	}
}
func WithClientSet(clientSet clientset.Interface) Option {
	return func(o *frameworkOptions) {
		o.clientSet = clientSet
	}
}
func WithKubeConfig(kubeConfig *restclient.Config) Option {
	return func(o *frameworkOptions) {
		o.kubeConfig = kubeConfig
	}
}
func WithInformerFactory(informerFactory informers.SharedInformerFactory) Option {
	return func(o *frameworkOptions) {
		o.informerFactory = informerFactory
	}
}
func WithSnapshotSharedLister(snapshotSharedLister framework.SharedLister) Option {
	return func(o *frameworkOptions) {
		o.snapshotSharedLister = snapshotSharedLister
	}
}
func WithRunAllFilters(runAllFilters bool) Option {
	return func(o *frameworkOptions) {
		o.runAllFilters = runAllFilters
	}
}
func WithEventRecorder(recorder events.EventRecorder) Option {
	return func(o *frameworkOptions) {
		o.eventRecorder = recorder
	}
}

type CaptureProfile func(config.KubeSchedulerProfile)

func WithCaptureProfile(c CaptureProfile) Option {
	return func(o *frameworkOptions) {
		o.captureProfile = c
	}
}
func WithClusterEventMap(m map[framework.ClusterEvent]sets.String) Option {
	return func(o *frameworkOptions) {
		o.clusterEventMap = m
	}
}
func WithParallelism(parallelism int) Option {
	return func(o *frameworkOptions) {
		o.parallelizer = parallelize.NewParallelizer(parallelism)
	}
}

type preemptHandle struct {
	framework.PodNominator
	framework.PluginsRunner
}

// RecorderFactory builds an EventRecorder for a given scheduler name.
type RecorderFactory func(string) events.EventRecorder

func NewRecorderFactory(b events.EventBroadcaster) RecorderFactory {
	return func(name string) events.EventRecorder {
		return b.NewRecorder(scheme.Scheme, name)
	}
}

// FrameworkFactory builds a Framework for a given profile configuration.
type FrameworkFactory func(config.KubeSchedulerProfile, ...Option) (Framework, error)

type Frameworks map[string]*Framework

func (profiles Frameworks) HandlesSchedulerName(name string) bool {
	_, ok := profiles[name]
	return ok
}

func NewFrameworks(profiles []config.KubeSchedulerProfile, r Registry,
	recorderFact RecorderFactory, opts ...Option) (Frameworks, error) {
	frameworks := make(Frameworks)
	for _, profile := range profiles {
		recorder := recorderFact(profile.SchedulerName)
		opts = append(opts, WithEventRecorder(recorder))
		p, err := NewFramework(r, &profile, opts...)
		if err != nil {
			return nil, fmt.Errorf("creating profile for scheduler name %s: %v", profile.SchedulerName, err)
		}
		frameworks[profile.SchedulerName] = p
	}

	return frameworks, nil
}

type Framework struct {
	registry              Registry
	snapshotSharedLister  framework.SharedLister
	waitingPods           *waitingPodsMap
	pluginNameToWeightMap map[string]int
	queueSortPlugins      []framework.QueueSortPlugin
	preFilterPlugins      []framework.PreFilterPlugin
	filterPlugins         []framework.FilterPlugin
	postFilterPlugins     []framework.PostFilterPlugin
	preScorePlugins       []framework.PreScorePlugin
	scorePlugins          []framework.ScorePlugin
	reservePlugins        []framework.ReservePlugin
	preBindPlugins        []framework.PreBindPlugin
	bindPlugins           []framework.BindPlugin
	postBindPlugins       []framework.PostBindPlugin
	permitPlugins         []framework.PermitPlugin

	clientSet       clientset.Interface
	eventRecorder   events.EventRecorder
	informerFactory informers.SharedInformerFactory

	metricsRecorder *metricsRecorder
	profileName     string

	preemptHandle framework.PreemptHandle

	// Indicates that RunFilterPlugins should accumulate all failed statuses and not return
	// after the first failure.
	runAllFilters bool
}

func NewFramework(r Registry, profile *config.KubeSchedulerProfile, opts ...Option) (*Framework, error) {
	options := defaultFrameworkOptions
	for _, opt := range opts {
		opt(&options)
	}

	f := &Framework{
		registry:              r,
		snapshotSharedLister:  options.snapshotSharedLister,
		pluginNameToWeightMap: make(map[string]int),
		waitingPods:           newWaitingPodsMap(),
		clientSet:             options.clientSet,
		eventRecorder:         options.eventRecorder,
		informerFactory:       options.informerFactory,
		metricsRecorder:       options.metricsRecorder,
		profileName:           options.profileName,
		runAllFilters:         options.runAllFilters,
	}
	f.preemptHandle = &preemptHandle{
		extenders:     options.extenders,
		PodNominator:  options.podNominator,
		PluginsRunner: f,
	}
	if plugins == nil {
		return f, nil
	}

	// get needed plugins from config
	pg := f.pluginsNeeded(plugins)

	pluginConfig := make(map[string]runtime.Object, len(args))
	for i := range args {
		name := args[i].Name
		if _, ok := pluginConfig[name]; ok {
			return nil, fmt.Errorf("repeated config for plugin %s", name)
		}
		pluginConfig[name] = args[i].Args
	}

	pluginsMap := make(map[string]framework.Plugin)
	var totalPriority int64
	for name, factory := range r {
		// initialize only needed plugins.
		if _, ok := pg[name]; !ok {
			continue
		}

		args, err := getPluginArgsOrDefault(pluginConfig, name)
		if err != nil {
			return nil, fmt.Errorf("getting args for Plugin %q: %w", name, err)
		}
		p, err := factory(args, f)
		if err != nil {
			return nil, fmt.Errorf("error initializing plugin %q: %v", name, err)
		}
		pluginsMap[name] = p

		// a weight of zero is not permitted, plugins can be disabled explicitly
		// when configured.
		f.pluginNameToWeightMap[name] = int(pg[name].Weight)
		if f.pluginNameToWeightMap[name] == 0 {
			f.pluginNameToWeightMap[name] = 1
		}
		// Checks totalPriority against MaxTotalScore to avoid overflow
		if int64(f.pluginNameToWeightMap[name])*framework.MaxNodeScore > framework.MaxTotalScore-totalPriority {
			return nil, fmt.Errorf("total score of Score plugins could overflow")
		}
		totalPriority += int64(f.pluginNameToWeightMap[name]) * framework.MaxNodeScore
	}

	for _, e := range f.getExtensionPoints(plugins) {
		if err := updatePluginList(e.slicePtr, e.plugins, pluginsMap); err != nil {
			return nil, err
		}
	}

	// Verifying the score weights again since Plugin.Name() could return a different
	// value from the one used in the configuration.
	for _, scorePlugin := range f.scorePlugins {
		if f.pluginNameToWeightMap[scorePlugin.Name()] == 0 {
			return nil, fmt.Errorf("score plugin %q is not configured with weight", scorePlugin.Name())
		}
	}

	if len(f.queueSortPlugins) == 0 {
		return nil, fmt.Errorf("no queue sort plugin is enabled")
	}
	if len(f.queueSortPlugins) > 1 {
		return nil, fmt.Errorf("only one queue sort plugin can be enabled")
	}
	if len(f.bindPlugins) == 0 {
		return nil, fmt.Errorf("at least one bind plugin is needed")
	}

	return f, nil
}

// SnapshotSharedLister returns the scheduler's SharedLister of the latest NodeInfo
// snapshot. The snapshot is taken at the beginning of a scheduling cycle and remains
// unchanged until a pod finishes "Reserve". There is no guarantee that the information
// remains unchanged after "Reserve".
func (f *Framework) SnapshotSharedLister() framework.SharedLister {
	return f.snapshotSharedLister
}

func (f *Framework) IterateOverWaitingPods(callback func(framework.WaitingPod)) {
	panic("implement me")
}

func (f *Framework) GetWaitingPod(uid types.UID) framework.WaitingPod {
	panic("implement me")
}

func (f *Framework) RejectWaitingPod(uid types.UID) {
	panic("implement me")
}

func (f *Framework) ClientSet() clientset.Interface {
	panic("implement me")
}

func (f *Framework) EventRecorder() events.EventRecorder {
	panic("implement me")
}

func (f *Framework) SharedInformerFactory() informers.SharedInformerFactory {
	panic("implement me")
}

func (f *Framework) PreemptHandle() framework.PreemptHandle {
	panic("implement me")
}

func (f *Framework) QueueSortFunc() framework.LessFunc {
	if f == nil {
		// If Framework is nil, simply keep their order unchanged.
		// NOTE: this is primarily for tests.
		return func(_, _ *framework.QueuedPodInfo) bool { return false }
	}

	if len(f.queueSortPlugins) == 0 {
		panic("No QueueSort plugin is registered in the Framework.")
	}

	// Only one QueueSort plugin can be enabled.
	return f.queueSortPlugins[0].Less
}

func (f *Framework) RunPreFilterPlugins(ctx context.Context, state *framework.CycleState, pod *v1.Pod) *framework.Status {
	panic("implement me")
}

func (f *Framework) RunFilterPlugins(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeInfo *framework.NodeInfo) framework.PluginToStatus {
	panic("implement me")
}

// INFO: pod在当前调度周期 filter extension point 失败时，执行抢占preemption逻辑，但是在下一个调度周期再去执行调度
func (f *Framework) RunPostFilterPlugins(ctx context.Context, state *framework.CycleState, pod *v1.Pod,
	filteredNodeStatusMap framework.NodeToStatusMap) (_ *framework.PostFilterResult, status *framework.Status) {
	startTime := time.Now()
	defer func() {
		metrics.FrameworkExtensionPointDuration.WithLabelValues(postFilter, status.Code().String(),
			f.profileName).Observe(metrics.SinceInSeconds(startTime))
	}()

	statuses := make(framework.PluginToStatus)
	for _, pl := range f.postFilterPlugins {
		r, s := f.runPostFilterPlugin(ctx, pl, state, pod, filteredNodeStatusMap)
		if s.IsSuccess() {
			return r, s
		} else if !s.IsUnschedulable() {
			// Any status other than Success or Unschedulable is Error.
			return nil, framework.NewStatus(framework.Error, s.Message())
		}

		statuses[pl.Name()] = s
	}

	return nil, statuses.Merge()
}

func (f *Framework) runPostFilterPlugin(ctx context.Context, pl framework.PostFilterPlugin, state *framework.CycleState, pod *v1.Pod, filteredNodeStatusMap framework.NodeToStatusMap) (*framework.PostFilterResult, *framework.Status) {
	if !state.ShouldRecordPluginMetrics() { // INFO: 有90%概率不需要record plugin metrics
		return pl.PostFilter(ctx, state, pod, filteredNodeStatusMap)
	}

	startTime := time.Now()
	r, s := pl.PostFilter(ctx, state, pod, filteredNodeStatusMap)
	f.metricsRecorder.observePluginDurationAsync(postFilter, pl.Name(), s, metrics.SinceInSeconds(startTime))
	return r, s
}

func (f *Framework) RunPreFilterExtensionAddPod(ctx context.Context, state *framework.CycleState, podToSchedule *v1.Pod, podToAdd *v1.Pod, nodeInfo *framework.NodeInfo) *framework.Status {
	panic("implement me")
}

func (f *Framework) RunPreFilterExtensionRemovePod(ctx context.Context, state *framework.CycleState, podToSchedule *v1.Pod, podToAdd *v1.Pod, nodeInfo *framework.NodeInfo) *framework.Status {
	panic("implement me")
}

func (f *Framework) RunPreScorePlugins(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodes []*v1.Node) *framework.Status {
	panic("implement me")
}

func (f *Framework) RunScorePlugins(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodes []*v1.Node) (framework.PluginToNodeScores, *framework.Status) {
	panic("implement me")
}

func (f *Framework) RunPreBindPlugins(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeName string) *framework.Status {
	panic("implement me")
}

func (f *Framework) RunPostBindPlugins(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeName string) {
	panic("implement me")
}

func (f *Framework) RunReservePluginsReserve(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeName string) *framework.Status {
	panic("implement me")
}

func (f *Framework) RunReservePluginsUnreserve(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeName string) {
	panic("implement me")
}

func (f *Framework) RunPermitPlugins(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeName string) *framework.Status {
	panic("implement me")
}

func (f *Framework) WaitOnPermit(ctx context.Context, pod *v1.Pod) *framework.Status {
	panic("implement me")
}

func (f *Framework) RunBindPlugins(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeName string) *framework.Status {
	panic("implement me")
}

func (f *Framework) HasFilterPlugins() bool {
	panic("implement me")
}

func (f *Framework) HasPostFilterPlugins() bool {
	return len(f.postFilterPlugins) > 0
}

func (f *Framework) HasScorePlugins() bool {
	panic("implement me")
}

func (f *Framework) ListPlugins() map[string][]config.Plugin {
	panic("implement me")
}

func (f *Framework) pluginsNeeded(plugins *config.Plugins) map[string]config.Plugin {
	pgMap := make(map[string]config.Plugin)

	if plugins == nil {
		return pgMap
	}

	find := func(pgs *config.PluginSet) {
		if pgs == nil {
			return
		}
		for _, pg := range pgs.Enabled {
			pgMap[pg.Name] = pg
		}
	}
	for _, e := range f.getExtensionPoints(plugins) {
		find(e.plugins)
	}

	return pgMap
}

// extensionPoint encapsulates desired and applied set of plugins at a specific extension
// point. This is used to simplify iterating over all extension points supported by the
// Framework.
type extensionPoint struct {
	// the set of plugins to be configured at this extension point.
	plugins *config.PluginSet
	// a pointer to the slice storing plugins implementations that will run at this
	// extension point.
	slicePtr interface{}
}

func (f *Framework) getExtensionPoints(plugins *config.Plugins) []extensionPoint {
	return []extensionPoint{
		{plugins.PreFilter, &f.preFilterPlugins},
		{plugins.Filter, &f.filterPlugins},
		{plugins.PostFilter, &f.postFilterPlugins},
		{plugins.Reserve, &f.reservePlugins},
		{plugins.PreScore, &f.preScorePlugins},
		{plugins.Score, &f.scorePlugins},
		{plugins.PreBind, &f.preBindPlugins},
		{plugins.Bind, &f.bindPlugins},
		{plugins.PostBind, &f.postBindPlugins},
		{plugins.Permit, &f.permitPlugins},
		{plugins.QueueSort, &f.queueSortPlugins},
	}
}

var defaultFrameworkOptions = frameworkOptions{
	metricsRecorder: newMetricsRecorder(1000, time.Second),
}

// NewFramework INFO: Framework 对象是一个框架对象，管理多个hooks点，在每一个hook点运行多个plugins

func updatePluginList(pluginList interface{}, pluginSet *config.PluginSet, pluginsMap map[string]framework.Plugin) error {
	return nil
}

var configDecoder = scheme.Codecs.UniversalDecoder()

// getPluginArgsOrDefault returns a configuration provided by the user or builds
// a default from the scheme. Returns `nil, nil` if the plugin does not have a
// defined arg types, such as in-tree plugins that don't require configuration
// or out-of-tree plugins.
func getPluginArgsOrDefault(pluginConfig map[string]runtime.Object, name string) (runtime.Object, error) {
	res, ok := pluginConfig[name]
	if ok {
		return res, nil
	}
	// Use defaults from latest config API version.
	gvk := schedulerapiv1.SchemeGroupVersion.WithKind(name + "Args")
	obj, _, err := configDecoder.Decode(nil, &gvk, nil)
	if runtime.IsNotRegisteredError(err) {
		// This plugin is out-of-tree or doesn't require configuration.
		return nil, nil
	}
	return obj, err
}
