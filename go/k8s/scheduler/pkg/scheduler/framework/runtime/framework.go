package runtime

import (
	"context"
	"fmt"
	"time"

	"k8s-lx1036/k8s/scheduler/pkg/scheduler/apis/config"
	"k8s-lx1036/k8s/scheduler/pkg/scheduler/apis/config/scheme"
	schedulerapiv1 "k8s-lx1036/k8s/scheduler/pkg/scheduler/apis/config/v1"
	framework "k8s-lx1036/k8s/scheduler/pkg/scheduler/framework/v1alpha1"
	"k8s-lx1036/k8s/scheduler/pkg/scheduler/metrics"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/informers"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/events"
)

const (
	postFilter = "PostFilter"
)

type frameworkOptions struct {
	clientSet            clientset.Interface
	eventRecorder        events.EventRecorder
	informerFactory      informers.SharedInformerFactory
	snapshotSharedLister framework.SharedLister
	metricsRecorder      *metricsRecorder
	profileName          string
	podNominator         framework.PodNominator
	extenders            []framework.Extender
	runAllFilters        bool
}

// Option for the frameworkImpl.
type Option func(*frameworkOptions)

// WithPodNominator sets podNominator for the scheduling frameworkImpl.
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

type preemptHandle struct {
	extenders []framework.Extender
	framework.PodNominator
	framework.PluginsRunner
}

// Extenders returns the registered extenders.
func (ph *preemptHandle) Extenders() []framework.Extender {
	return ph.extenders
}

// INFO: frameworkImpl 对象有点像 iptables
type frameworkImpl struct {
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

// SnapshotSharedLister returns the scheduler's SharedLister of the latest NodeInfo
// snapshot. The snapshot is taken at the beginning of a scheduling cycle and remains
// unchanged until a pod finishes "Reserve". There is no guarantee that the information
// remains unchanged after "Reserve".
func (f *frameworkImpl) SnapshotSharedLister() framework.SharedLister {
	return f.snapshotSharedLister
}

func (f *frameworkImpl) IterateOverWaitingPods(callback func(framework.WaitingPod)) {
	panic("implement me")
}

func (f *frameworkImpl) GetWaitingPod(uid types.UID) framework.WaitingPod {
	panic("implement me")
}

func (f *frameworkImpl) RejectWaitingPod(uid types.UID) {
	panic("implement me")
}

func (f *frameworkImpl) ClientSet() clientset.Interface {
	panic("implement me")
}

func (f *frameworkImpl) EventRecorder() events.EventRecorder {
	panic("implement me")
}

func (f *frameworkImpl) SharedInformerFactory() informers.SharedInformerFactory {
	panic("implement me")
}

func (f *frameworkImpl) PreemptHandle() framework.PreemptHandle {
	panic("implement me")
}

func (f *frameworkImpl) QueueSortFunc() framework.LessFunc {
	if f == nil {
		// If frameworkImpl is nil, simply keep their order unchanged.
		// NOTE: this is primarily for tests.
		return func(_, _ *framework.QueuedPodInfo) bool { return false }
	}

	if len(f.queueSortPlugins) == 0 {
		panic("No QueueSort plugin is registered in the frameworkImpl.")
	}

	// Only one QueueSort plugin can be enabled.
	return f.queueSortPlugins[0].Less
}

func (f *frameworkImpl) RunPreFilterPlugins(ctx context.Context, state *framework.CycleState, pod *v1.Pod) *framework.Status {
	panic("implement me")
}

func (f *frameworkImpl) RunFilterPlugins(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeInfo *framework.NodeInfo) framework.PluginToStatus {
	panic("implement me")
}

// INFO: pod在当前调度周期 filter extension point 失败时，执行抢占preemption逻辑，但是在下一个调度周期再去执行调度
func (f *frameworkImpl) RunPostFilterPlugins(ctx context.Context, state *framework.CycleState, pod *v1.Pod,
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

func (f *frameworkImpl) runPostFilterPlugin(ctx context.Context, pl framework.PostFilterPlugin, state *framework.CycleState, pod *v1.Pod, filteredNodeStatusMap framework.NodeToStatusMap) (*framework.PostFilterResult, *framework.Status) {
	if !state.ShouldRecordPluginMetrics() { // INFO: 有90%概率不需要record plugin metrics
		return pl.PostFilter(ctx, state, pod, filteredNodeStatusMap)
	}

	startTime := time.Now()
	r, s := pl.PostFilter(ctx, state, pod, filteredNodeStatusMap)
	f.metricsRecorder.observePluginDurationAsync(postFilter, pl.Name(), s, metrics.SinceInSeconds(startTime))
	return r, s
}

func (f *frameworkImpl) RunPreFilterExtensionAddPod(ctx context.Context, state *framework.CycleState, podToSchedule *v1.Pod, podToAdd *v1.Pod, nodeInfo *framework.NodeInfo) *framework.Status {
	panic("implement me")
}

func (f *frameworkImpl) RunPreFilterExtensionRemovePod(ctx context.Context, state *framework.CycleState, podToSchedule *v1.Pod, podToAdd *v1.Pod, nodeInfo *framework.NodeInfo) *framework.Status {
	panic("implement me")
}

func (f *frameworkImpl) RunPreScorePlugins(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodes []*v1.Node) *framework.Status {
	panic("implement me")
}

func (f *frameworkImpl) RunScorePlugins(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodes []*v1.Node) (framework.PluginToNodeScores, *framework.Status) {
	panic("implement me")
}

func (f *frameworkImpl) RunPreBindPlugins(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeName string) *framework.Status {
	panic("implement me")
}

func (f *frameworkImpl) RunPostBindPlugins(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeName string) {
	panic("implement me")
}

func (f *frameworkImpl) RunReservePluginsReserve(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeName string) *framework.Status {
	panic("implement me")
}

func (f *frameworkImpl) RunReservePluginsUnreserve(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeName string) {
	panic("implement me")
}

func (f *frameworkImpl) RunPermitPlugins(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeName string) *framework.Status {
	panic("implement me")
}

func (f *frameworkImpl) WaitOnPermit(ctx context.Context, pod *v1.Pod) *framework.Status {
	panic("implement me")
}

func (f *frameworkImpl) RunBindPlugins(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeName string) *framework.Status {
	panic("implement me")
}

func (f *frameworkImpl) HasFilterPlugins() bool {
	panic("implement me")
}

func (f *frameworkImpl) HasPostFilterPlugins() bool {
	return len(f.postFilterPlugins) > 0
}

func (f *frameworkImpl) HasScorePlugins() bool {
	panic("implement me")
}

func (f *frameworkImpl) ListPlugins() map[string][]config.Plugin {
	panic("implement me")
}

func (f *frameworkImpl) pluginsNeeded(plugins *config.Plugins) map[string]config.Plugin {
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
// frameworkImpl.
type extensionPoint struct {
	// the set of plugins to be configured at this extension point.
	plugins *config.PluginSet
	// a pointer to the slice storing plugins implementations that will run at this
	// extension point.
	slicePtr interface{}
}

func (f *frameworkImpl) getExtensionPoints(plugins *config.Plugins) []extensionPoint {
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
func NewFramework(r Registry, plugins *config.Plugins, args []config.PluginConfig, opts ...Option) (framework.Framework, error) {
	options := defaultFrameworkOptions
	for _, opt := range opts {
		opt(&options)
	}

	f := &frameworkImpl{
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
