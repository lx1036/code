package runtime

import (
	"context"
	"fmt"
	internalqueue "k8s-lx1036/k8s/scheduler/pkg/internal/queue"
	"k8s.io/kubernetes/pkg/scheduler/metrics"
	"reflect"
	"time"

	"k8s-lx1036/k8s/scheduler/pkg/apis/config/scheme"
	configv1 "k8s-lx1036/k8s/scheduler/pkg/apis/config/v1"
	"k8s-lx1036/k8s/scheduler/pkg/framework"
	"k8s-lx1036/k8s/scheduler/pkg/framework/parallelize"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/informers"
	clientset "k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/events"
	corev1helpers "k8s.io/component-helpers/scheduling/corev1"
	"k8s.io/klog/v2"
)

// TODO: record plugin metrics

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
	podNominator         *internalqueue.PodNominator
	runAllFilters        bool
	captureProfile       CaptureProfile
	clusterEventMap      map[framework.ClusterEvent]sets.String
	parallelizer         parallelize.Parallelizer
}

// Option for the Framework.
type Option func(*frameworkOptions)

// WithPodNominator sets podNominator for the scheduling Framework.
func WithPodNominator(nominator *internalqueue.PodNominator) Option {
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

type CaptureProfile func(configv1.KubeSchedulerProfile)

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

// RecorderFactory builds an EventRecorder for a given scheduler name.
type RecorderFactory func(string) events.EventRecorder

func NewRecorderFactory(b events.EventBroadcaster) RecorderFactory {
	return func(name string) events.EventRecorder {
		return b.NewRecorder(scheme.Scheme, name)
	}
}

// FrameworkFactory builds a Framework for a given profile configuration.
type FrameworkFactory func(configv1.KubeSchedulerProfile, ...Option) (Framework, error)

type Frameworks map[string]*Framework

func (profiles Frameworks) HandlesSchedulerName(name string) bool {
	_, ok := profiles[name]
	return ok
}

func NewFrameworks(profiles []configv1.KubeSchedulerProfile, r Registry,
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
	waitingPods           *WaitingPodsMap
	pluginNameToWeightMap map[string]int
	queueSortPlugins      []framework.QueueSortPlugin
	preFilterPlugins      []framework.PreFilterPlugin
	filterPlugins         []framework.FilterPlugin
	postFilterPlugins     []framework.PostFilterPlugin
	preScorePlugins       []framework.PreScorePlugin

	scorePlugins      []framework.ScorePlugin
	scorePluginWeight map[string]int

	reservePlugins  []framework.ReservePlugin
	preBindPlugins  []framework.PreBindPlugin
	bindPlugins     []framework.BindPlugin
	postBindPlugins []framework.PostBindPlugin
	permitPlugins   []framework.PermitPlugin

	kubeConfig      *restclient.Config
	clientSet       clientset.Interface
	eventRecorder   events.EventRecorder
	informerFactory informers.SharedInformerFactory

	metricsRecorder *metricsRecorder
	profileName     string

	preemptHandle framework.PreemptHandle
	PodNominator  *internalqueue.PodNominator

	parallelizer parallelize.Parallelizer

	// Indicates that RunFilterPlugins should accumulate all failed statuses and not return
	// after the first failure.
	runAllFilters bool
}

var defaultFrameworkOptions = frameworkOptions{
	metricsRecorder: newMetricsRecorder(1000, time.Second),
	clusterEventMap: make(map[framework.ClusterEvent]sets.String),
	parallelizer:    parallelize.NewParallelizer(parallelize.DefaultParallelism),
}

// NewFramework 该函数实例化了一个 profile 包含的所有 plugins
func NewFramework(r Registry, profile *configv1.KubeSchedulerProfile, opts ...Option) (*Framework, error) {
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

		scorePluginWeight: make(map[string]int),
		kubeConfig:        options.kubeConfig,
		PodNominator:      options.podNominator,

		parallelizer: options.parallelizer,
	}
	if profile == nil {
		return f, nil
	}
	f.profileName = profile.SchedulerName
	if profile.Plugins == nil {
		return f, nil
	}

	pluginConfig := make(map[string]runtime.Object, len(profile.PluginConfig))
	for i := range profile.PluginConfig {
		name := profile.PluginConfig[i].Name
		if _, ok := pluginConfig[name]; ok {
			return nil, fmt.Errorf("repeated config for plugin %s", name)
		}
		pluginConfig[name] = profile.PluginConfig[i].Args
	}

	pluginsMap := make(map[string]framework.Plugin)
	neededPlugins := f.pluginsNeeded(profile.Plugins) // get needed plugins from config file
	outputProfile := configv1.KubeSchedulerProfile{
		SchedulerName: f.profileName,
		Plugins:       profile.Plugins,
		PluginConfig:  make([]configv1.PluginConfig, 0, len(neededPlugins)),
	}
	for name, pluginFactory := range r { // r 包含所有 in-tree and out-of-tree plugins
		// initialize only needed plugins.
		if _, ok := neededPlugins[name]; !ok {
			continue
		}

		args := pluginConfig[name] // merge plugin args
		if args != nil {
			outputProfile.PluginConfig = append(outputProfile.PluginConfig, configv1.PluginConfig{
				Name: name,
				Args: args,
			})
		}

		p, err := pluginFactory(args, f)
		if err != nil {
			return nil, fmt.Errorf("error initializing plugin %q: %v", name, err)
		}
		pluginsMap[name] = p

		// Update ClusterEventMap in place.
		fillEventToPluginMap(p, options.clusterEventMap)
	}

	// INFO: 这里函数会按照一个个 hook 点去从 profile.Plugins 去实例化 framework 的一堆 plugins，包括必备的 queueSort 和 bind plugins
	//  这里不追 updatePluginList() 函数的逻辑了，节省时间，直接复制
	for _, e := range f.getExtensionPoints(profile.Plugins) {
		if err := updatePluginList(e.slicePtr, *e.plugins, pluginsMap); err != nil {
			return nil, err
		}
	}
	if len(f.queueSortPlugins) != 1 {
		return nil, fmt.Errorf("one queue sort plugin required for profile with scheduler name %q", profile.SchedulerName)
	}
	if len(f.bindPlugins) == 0 {
		return nil, fmt.Errorf("at least one bind plugin is needed")
	}

	if err := f.getScoreWeights(pluginsMap, profile.Plugins.Score.Enabled); err != nil {
		return nil, err
	}

	return f, nil
}

func updatePluginList(pluginList interface{}, pluginSet configv1.PluginSet, pluginsMap map[string]framework.Plugin) error {
	plugins := reflect.ValueOf(pluginList).Elem()
	pluginType := plugins.Type().Elem()
	set := sets.NewString()
	for _, ep := range pluginSet.Enabled {
		pg, ok := pluginsMap[ep.Name]
		if !ok {
			return fmt.Errorf("%s %q does not exist", pluginType.Name(), ep.Name)
		}

		if !reflect.TypeOf(pg).Implements(pluginType) {
			return fmt.Errorf("plugin %q does not extend %s plugin", ep.Name, pluginType.Name())
		}

		if set.Has(ep.Name) {
			return fmt.Errorf("plugin %q already registered as %q", ep.Name, pluginType.Name())
		}

		set.Insert(ep.Name)

		newPlugins := reflect.Append(plugins, reflect.ValueOf(pg))
		plugins.Set(newPlugins)
	}
	return nil
}

func fillEventToPluginMap(p framework.Plugin, eventToPlugins map[framework.ClusterEvent]sets.String) {

}

func (f *Framework) Parallelizer() parallelize.Parallelizer {
	return f.parallelizer
}

func (f *Framework) getScoreWeights(pluginsMap map[string]framework.Plugin, plugins []configv1.Plugin) error {
	var totalPriority int64
	for _, e := range plugins {

		if _, ok := f.scorePluginWeight[e.Name]; ok {
			continue
		}
		f.scorePluginWeight[e.Name] = int(e.Weight)
		if f.scorePluginWeight[e.Name] == 0 {
			f.scorePluginWeight[e.Name] = 1
		}

		// Checks totalPriority against MaxTotalScore to avoid overflow
		if int64(f.scorePluginWeight[e.Name])*framework.MaxNodeScore > framework.MaxTotalScore-totalPriority {
			return fmt.Errorf("total score of Score plugins could overflow")
		}
		totalPriority += int64(f.scorePluginWeight[e.Name]) * framework.MaxNodeScore
	}

	for _, scorePlugin := range f.scorePlugins {
		if f.scorePluginWeight[scorePlugin.Name()] == 0 {
			return fmt.Errorf("score plugin %q is not configured with weight", scorePlugin.Name())
		}
	}

	return nil
}

func (f *Framework) pluginsNeeded(plugins *configv1.Plugins) map[string]configv1.Plugin {
	pgMap := make(map[string]configv1.Plugin)
	if plugins == nil {
		return pgMap
	}

	find := func(pgs *configv1.PluginSet) {
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

type extensionPoint struct {
	// the set of plugins to be configured at this extension point.
	plugins *configv1.PluginSet
	// a pointer to the slice storing plugins implementations that will run at this
	// extension point.
	slicePtr interface{}
}

// INFO: 这里是 framework 的排序的 hook 点
func (f *Framework) getExtensionPoints(plugins *configv1.Plugins) []extensionPoint {
	return []extensionPoint{
		{&plugins.QueueSort, &f.queueSortPlugins},

		{&plugins.PreFilter, &f.preFilterPlugins},
		{&plugins.Filter, &f.filterPlugins},
		{&plugins.PostFilter, &f.postFilterPlugins},

		{&plugins.Reserve, &f.reservePlugins},
		{&plugins.PreScore, &f.preScorePlugins},
		{&plugins.Score, &f.scorePlugins},

		{&plugins.PreBind, &f.preBindPlugins},
		{&plugins.Bind, &f.bindPlugins},
		{&plugins.PostBind, &f.postBindPlugins},

		{&plugins.Permit, &f.permitPlugins},
	}
}

func (f *Framework) SnapshotSharedLister() framework.SharedLister {
	return f.snapshotSharedLister
}

func (f *Framework) IterateOverWaitingPods(callback func(*WaitingPod)) {
	f.waitingPods.iterate(callback)
}

func (f *Framework) GetWaitingPod(uid types.UID) *WaitingPod {
	if wp := f.waitingPods.get(uid); wp != nil {
		return wp
	}
	return nil // Returning nil instead of *waitingPod(nil).
}

func (f *Framework) RejectWaitingPod(uid types.UID) {
	panic("implement me")
}

func (f *Framework) ProfileName() string {
	return f.profileName
}

func (f *Framework) ClientSet() clientset.Interface {
	return f.clientSet
}

func (f *Framework) KubeConfig() *restclient.Config {
	return f.kubeConfig
}

func (f *Framework) EventRecorder() events.EventRecorder {
	return f.eventRecorder
}

func (f *Framework) SharedInformerFactory() informers.SharedInformerFactory {
	return f.informerFactory
}

func (f *Framework) ListPlugins() *configv1.Plugins {
	m := configv1.Plugins{}
	for _, e := range f.getExtensionPoints(&m) {
		plugins := reflect.ValueOf(e.slicePtr).Elem()
		extName := plugins.Type().Elem().Name()
		var cfgs []configv1.Plugin
		for i := 0; i < plugins.Len(); i++ {
			name := plugins.Index(i).Interface().(framework.Plugin).Name()
			p := configv1.Plugin{Name: name}
			if extName == "ScorePlugin" {
				// Weights apply only to score plugins.
				p.Weight = int32(f.scorePluginWeight[name])
			}
			cfgs = append(cfgs, p)
		}
		if len(cfgs) > 0 {
			e.plugins.Enabled = cfgs
		}
	}
	return &m
}

func (f *Framework) QueueSortFunc() framework.LessFunc {
	if f == nil {
		// If Framework is nil, simply keep their order unchanged.
		// NOTE: this is primarily for tests.
		return func(_, _ *framework.QueuedPodInfo) bool { return false }
	}

	// Only one QueueSort plugin can be enabled.
	return f.queueSortPlugins[0].Less
}

func (f *Framework) HasFilterPlugins() bool {
	return len(f.filterPlugins) != 0
}

func (f *Framework) HasPostFilterPlugins() bool {
	return len(f.postFilterPlugins) > 0
}

func (f *Framework) RunPreFilterPlugins(ctx context.Context,
	state *framework.CycleState, pod *corev1.Pod) (_ *framework.PreFilterResult, status *framework.Status) {
	var result *framework.PreFilterResult
	var pluginsWithNodes []string
	for _, pl := range f.preFilterPlugins {
		r, s := f.runPreFilterPlugin(ctx, pl, state, pod)
		if !s.IsSuccess() {
			s.SetFailedPlugin(pl.Name())
			if s.IsUnschedulable() {
				return nil, s
			}
			return nil, framework.AsStatus(fmt.Errorf("running PreFilter plugin %q: %w", pl.Name(), status.AsError())).
				WithFailedPlugin(pl.Name())
		}

		if !r.AllNodes() {
			pluginsWithNodes = append(pluginsWithNodes, pl.Name())
		}
		result = result.Merge(r)
		if !result.AllNodes() && len(result.NodeNames) == 0 {
			msg := fmt.Sprintf("node(s) didn't satisfy plugin(s) %v simultaneously", pluginsWithNodes)
			if len(pluginsWithNodes) == 1 {
				msg = fmt.Sprintf("node(s) didn't satisfy plugin %v", pluginsWithNodes[0])
			}
			return nil, framework.NewStatus(framework.Unschedulable, msg)
		}
	}

	return result, nil
}

func (f *Framework) runPreFilterPlugin(ctx context.Context, pl framework.PreFilterPlugin,
	state *framework.CycleState, pod *corev1.Pod) (*framework.PreFilterResult, *framework.Status) {
	return pl.PreFilter(ctx, state, pod)
}

func (f *Framework) RunPreFilterExtensionAddPod(ctx context.Context, state *framework.CycleState, podToSchedule *corev1.Pod,
	podInfoToAdd *framework.PodInfo, nodeInfo *framework.NodeInfo) (status *framework.Status) {
	for _, pl := range f.preFilterPlugins {
		if pl.PreFilterExtensions() == nil {
			continue
		}
		status = f.runPreFilterExtensionAddPod(ctx, pl, state, podToSchedule, podInfoToAdd, nodeInfo)
		if !status.IsSuccess() {
			err := status.AsError()
			klog.ErrorS(err, "Failed running AddPod on PreFilter plugin", "plugin", pl.Name(), "pod", klog.KObj(podToSchedule))
			return framework.AsStatus(fmt.Errorf("running AddPod on PreFilter plugin %q: %w", pl.Name(), err))
		}
	}

	return nil
}
func (f *Framework) runPreFilterExtensionAddPod(ctx context.Context, pl framework.PreFilterPlugin, state *framework.CycleState,
	podToSchedule *corev1.Pod, podInfoToAdd *framework.PodInfo, nodeInfo *framework.NodeInfo) *framework.Status {
	return pl.PreFilterExtensions().AddPod(ctx, state, podToSchedule, podInfoToAdd, nodeInfo)
}
func (f *Framework) RunPreFilterExtensionRemovePod(ctx context.Context, state *framework.CycleState, podToSchedule *corev1.Pod,
	podInfoToRemove *framework.PodInfo, nodeInfo *framework.NodeInfo) (status *framework.Status) {
	for _, pl := range f.preFilterPlugins {
		if pl.PreFilterExtensions() == nil {
			continue
		}
		status = f.runPreFilterExtensionRemovePod(ctx, pl, state, podToSchedule, podInfoToRemove, nodeInfo)
		if !status.IsSuccess() {
			err := status.AsError()
			klog.ErrorS(err, "Failed running RemovePod on PreFilter plugin", "plugin", pl.Name(), "pod", klog.KObj(podToSchedule))
			return framework.AsStatus(fmt.Errorf("running RemovePod on PreFilter plugin %q: %w", pl.Name(), err))
		}
	}

	return nil
}
func (f *Framework) runPreFilterExtensionRemovePod(ctx context.Context, pl framework.PreFilterPlugin, state *framework.CycleState,
	podToSchedule *corev1.Pod, podInfoToRemove *framework.PodInfo, nodeInfo *framework.NodeInfo) *framework.Status {
	return pl.PreFilterExtensions().RemovePod(ctx, state, podToSchedule, podInfoToRemove, nodeInfo)
}

// RunFilterPluginsWithNominatedPods 两处会调用: Schedule and Preempt.
func (f *Framework) RunFilterPluginsWithNominatedPods(ctx context.Context, state *framework.CycleState, pod *corev1.Pod,
	nodeInfo *framework.NodeInfo) *framework.Status {
	var status *framework.Status
	podsAdded := false
	for i := 0; i < 2; i++ {
		stateToUse := state
		nodeInfoToUse := nodeInfo
		if i == 0 {
			var err error
			podsAdded, stateToUse, nodeInfoToUse, err = f.addNominatedPods(ctx, pod, state, nodeInfo)
			if err != nil {
				return framework.AsStatus(err)
			}
		} else if !podsAdded || !status.IsSuccess() {
			break
		}

		// 如果 pod 通不过 Filter plugins
		statusMap := f.RunFilterPlugins(ctx, stateToUse, pod, nodeInfoToUse)
		status = statusMap.Merge()
		if !status.IsSuccess() && !status.IsUnschedulable() {
			return status
		}
	}

	return status
}

// add pods with equal or greater priority than target pod
func (f *Framework) addNominatedPods(ctx context.Context, pod *corev1.Pod, state *framework.CycleState,
	nodeInfo *framework.NodeInfo) (bool, *framework.CycleState, *framework.NodeInfo, error) {
	nominatedPodInfos := f.PodNominator.NominatedPodsForNode(nodeInfo.Node().Name)
	if len(nominatedPodInfos) == 0 {
		return false, state, nodeInfo, nil
	}

	nodeInfoOut := nodeInfo.Clone()
	stateOut := state.Clone()
	podsAdded := false
	for _, pi := range nominatedPodInfos {
		if corev1helpers.PodPriority(pi.Pod) >= corev1helpers.PodPriority(pod) && pi.Pod.UID != pod.UID {
			nodeInfoOut.AddPodInfo(pi)
			status := f.RunPreFilterExtensionAddPod(ctx, stateOut, pod, pi, nodeInfoOut)
			if !status.IsSuccess() {
				return false, state, nodeInfo, status.AsError()
			}
			podsAdded = true
		}
	}

	return podsAdded, stateOut, nodeInfoOut, nil
}

func (f *Framework) RunFilterPlugins(ctx context.Context, state *framework.CycleState, pod *corev1.Pod, nodeInfo *framework.NodeInfo) framework.PluginToStatus {
	statuses := make(framework.PluginToStatus)
	for _, pl := range f.filterPlugins {
		pluginStatus := f.runFilterPlugin(ctx, pl, state, pod, nodeInfo)
		if !pluginStatus.IsSuccess() {
			if !pluginStatus.IsUnschedulable() {
				// Filter plugins are not supposed to return any status other than
				// Success or Unschedulable.
				errStatus := framework.AsStatus(fmt.Errorf("running %q filter plugin: %w", pl.Name(), pluginStatus.AsError())).WithFailedPlugin(pl.Name())
				return map[string]*framework.Status{pl.Name(): errStatus}
			}
			pluginStatus.SetFailedPlugin(pl.Name())
			statuses[pl.Name()] = pluginStatus
		}
	}

	return statuses
}

func (f *Framework) runFilterPlugin(ctx context.Context, pl framework.FilterPlugin, state *framework.CycleState, pod *corev1.Pod, nodeInfo *framework.NodeInfo) *framework.Status {
	return pl.Filter(ctx, state, pod, nodeInfo)
}

// RunPostFilterPlugins INFO: pod在当前调度周期 filter extension point 失败时，执行抢占preemption逻辑，但是在下一个调度周期再去执行调度
func (f *Framework) RunPostFilterPlugins(ctx context.Context, state *framework.CycleState, pod *corev1.Pod,
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

func (f *Framework) runPostFilterPlugin(ctx context.Context, pl framework.PostFilterPlugin, state *framework.CycleState,
	pod *corev1.Pod, filteredNodeStatusMap framework.NodeToStatusMap) (*framework.PostFilterResult, *framework.Status) {
	return pl.PostFilter(ctx, state, pod, filteredNodeStatusMap)
}

func (f *Framework) HasScorePlugins() bool {
	return len(f.scorePlugins) != 0
}

func (f *Framework) RunPreScorePlugins(ctx context.Context, state *framework.CycleState, pod *corev1.Pod, nodes []*corev1.Node) *framework.Status {
	var status *framework.Status
	for _, pl := range f.preScorePlugins {
		status = f.runPreScorePlugin(ctx, pl, state, pod, nodes)
		if !status.IsSuccess() {
			return framework.AsStatus(fmt.Errorf("running PreScore plugin %q: %w", pl.Name(), status.AsError()))
		}
	}

	return nil
}
func (f *Framework) runPreScorePlugin(ctx context.Context, pl framework.PreScorePlugin, state *framework.CycleState, pod *corev1.Pod, nodes []*corev1.Node) *framework.Status {
	return pl.PreScore(ctx, state, pod, nodes)
}

func (f *Framework) RunScorePlugins(ctx context.Context, state *framework.CycleState, pod *corev1.Pod, nodes []*corev1.Node) (framework.PluginToNodeScores, *framework.Status) {
	pluginToNodeScores := make(framework.PluginToNodeScores, len(f.scorePlugins))
	for _, pl := range f.scorePlugins {
		pluginToNodeScores[pl.Name()] = make(framework.NodeScoreList, len(nodes))
	}
	ctx, cancel := context.WithCancel(ctx)
	errCh := parallelize.NewErrorChannel()
	// Run Score method for each node in parallel.
	f.Parallelizer().Until(ctx, len(nodes), func(index int) { // 总共有 16 个 worker
		for _, pl := range f.scorePlugins {
			nodeName := nodes[index].Name
			s, status := f.runScorePlugin(ctx, pl, state, pod, nodeName)
			if !status.IsSuccess() {
				err := fmt.Errorf("plugin %q failed with: %w", pl.Name(), status.AsError())
				errCh.SendErrorWithCancel(err, cancel)
				return
			}
			pluginToNodeScores[pl.Name()][index] = framework.NodeScore{
				Name:  nodeName,
				Score: s,
			}
		}
	})
	if err := errCh.ReceiveError(); err != nil {
		return nil, framework.AsStatus(fmt.Errorf("running Score plugins: %w", err))
	}

	// Run NormalizeScore method for each ScorePlugin in parallel.
	f.Parallelizer().Until(ctx, len(f.scorePlugins), func(index int) {
		pl := f.scorePlugins[index]
		nodeScoreList := pluginToNodeScores[pl.Name()]
		if pl.ScoreExtensions() == nil {
			return
		}
		status := f.runScoreExtension(ctx, pl, state, pod, nodeScoreList)
		if !status.IsSuccess() {
			err := fmt.Errorf("plugin %q failed with: %w", pl.Name(), status.AsError())
			errCh.SendErrorWithCancel(err, cancel)
			return
		}
	})
	if err := errCh.ReceiveError(); err != nil {
		return nil, framework.AsStatus(fmt.Errorf("running Normalize on Score plugins: %w", err))
	}

	// Apply score defaultWeights for each ScorePlugin in parallel.
	f.Parallelizer().Until(ctx, len(f.scorePlugins), func(index int) {
		pl := f.scorePlugins[index]
		// Score plugins' weight has been checked when they are initialized.
		weight := f.scorePluginWeight[pl.Name()]
		nodeScoreList := pluginToNodeScores[pl.Name()]

		for i, nodeScore := range nodeScoreList {
			// return error if score plugin returns invalid score.
			if nodeScore.Score > framework.MaxNodeScore || nodeScore.Score < framework.MinNodeScore {
				err := fmt.Errorf("plugin %q returns an invalid score %v, it should in the range of [%v, %v] after normalizing", pl.Name(), nodeScore.Score, framework.MinNodeScore, framework.MaxNodeScore)
				errCh.SendErrorWithCancel(err, cancel)
				return
			}
			nodeScoreList[i].Score = nodeScore.Score * int64(weight)
		}
	})
	if err := errCh.ReceiveError(); err != nil {
		return nil, framework.AsStatus(fmt.Errorf("applying score defaultWeights on Score plugins: %w", err))
	}

	return pluginToNodeScores, nil
}
func (f *Framework) runScorePlugin(ctx context.Context, pl framework.ScorePlugin, state *framework.CycleState,
	pod *corev1.Pod, nodeName string) (int64, *framework.Status) {
	return pl.Score(ctx, state, pod, nodeName)
}
func (f *Framework) runScoreExtension(ctx context.Context, pl framework.ScorePlugin, state *framework.CycleState,
	pod *corev1.Pod, nodeScoreList framework.NodeScoreList) *framework.Status {
	return pl.ScoreExtensions().NormalizeScore(ctx, state, pod, nodeScoreList)
}

func (f *Framework) RunReservePluginsReserve(ctx context.Context, state *framework.CycleState, pod *corev1.Pod, nodeName string) *framework.Status {
	var status *framework.Status
	for _, pl := range f.reservePlugins {
		status = f.runReservePluginReserve(ctx, pl, state, pod, nodeName)
		if !status.IsSuccess() {
			err := status.AsError()
			klog.ErrorS(err, "Failed running Reserve plugin", "plugin", pl.Name(), "pod", klog.KObj(pod))
			return framework.AsStatus(fmt.Errorf("running Reserve plugin %q: %w", pl.Name(), err))
		}
	}
	return nil
}
func (f *Framework) runReservePluginReserve(ctx context.Context, pl framework.ReservePlugin, state *framework.CycleState,
	pod *corev1.Pod, nodeName string) *framework.Status {
	return pl.Reserve(ctx, state, pod, nodeName)
}

func (f *Framework) RunReservePluginsUnreserve(ctx context.Context, state *framework.CycleState, pod *corev1.Pod, nodeName string) {
	// Execute the Unreserve operation of each reserve plugin in the
	// *reverse* order in which the Reserve operation was executed.
	for i := len(f.reservePlugins) - 1; i >= 0; i-- {
		f.runReservePluginUnreserve(ctx, f.reservePlugins[i], state, pod, nodeName)
	}
}
func (f *Framework) runReservePluginUnreserve(ctx context.Context, pl framework.ReservePlugin, state *framework.CycleState,
	pod *corev1.Pod, nodeName string) {
	pl.Unreserve(ctx, state, pod, nodeName)
}

const (
	maxTimeout = 15 * time.Minute
)

func (f *Framework) RunPermitPlugins(ctx context.Context, state *framework.CycleState, pod *corev1.Pod, nodeName string) *framework.Status {
	pluginsWaitTime := make(map[string]time.Duration)
	statusCode := framework.Success
	for _, pl := range f.permitPlugins {
		status, timeout := f.runPermitPlugin(ctx, pl, state, pod, nodeName)
		if !status.IsSuccess() {
			if status.IsUnschedulable() {
				klog.V(4).InfoS("Pod rejected by permit plugin", "pod", klog.KObj(pod), "plugin", pl.Name(), "status", status.Message())
				status.SetFailedPlugin(pl.Name())
				return status
			}
			if status.Code() == framework.Wait {
				// Not allowed to be greater than maxTimeout.
				if timeout > maxTimeout {
					timeout = maxTimeout
				}
				pluginsWaitTime[pl.Name()] = timeout
				statusCode = framework.Wait
			} else {
				err := status.AsError()
				klog.ErrorS(err, "Failed running Permit plugin", "plugin", pl.Name(), "pod", klog.KObj(pod))
				return framework.AsStatus(fmt.Errorf("running Permit plugin %q: %w", pl.Name(), err)).WithFailedPlugin(pl.Name())
			}
		}
	}

	if statusCode == framework.Wait {
		waitingPod := newWaitingPod(pod, pluginsWaitTime)
		f.waitingPods.add(waitingPod)
		msg := fmt.Sprintf("one or more plugins asked to wait and no plugin rejected pod %q", pod.Name)
		klog.V(4).InfoS("One or more plugins asked to wait and no plugin rejected pod", "pod", klog.KObj(pod))
		return framework.NewStatus(framework.Wait, msg)
	}
	return nil
}
func (f *Framework) runPermitPlugin(ctx context.Context, pl framework.PermitPlugin, state *framework.CycleState,
	pod *corev1.Pod, nodeName string) (*framework.Status, time.Duration) {
	return pl.Permit(ctx, state, pod, nodeName)
}

// WaitOnPermit will block, if the pod is a waiting pod, until the waiting pod is rejected or allowed.
func (f *Framework) WaitOnPermit(ctx context.Context, pod *corev1.Pod) *framework.Status {
	waitingPod := f.waitingPods.get(pod.UID)
	if waitingPod == nil {
		return nil
	}
	defer f.waitingPods.remove(pod.UID)

	status := <-waitingPod.status
	if !status.IsSuccess() {
		if status.IsUnschedulable() {
			klog.V(4).InfoS("Pod rejected while waiting on permit", "pod", klog.KObj(pod), "status", status.Message())
			status.SetFailedPlugin(status.FailedPlugin())
			return status
		}
		err := status.AsError()
		klog.ErrorS(err, "Failed waiting on permit for pod", "pod", klog.KObj(pod))
		return framework.AsStatus(fmt.Errorf("waiting on permit for pod: %w", err)).WithFailedPlugin(status.FailedPlugin())
	}
	return nil
}

func (f *Framework) RunPreBindPlugins(ctx context.Context, state *framework.CycleState, pod *corev1.Pod, nodeName string) *framework.Status {
	var status *framework.Status
	for _, pl := range f.preBindPlugins {
		status = f.runPreBindPlugin(ctx, pl, state, pod, nodeName)
		if !status.IsSuccess() {
			err := status.AsError()
			klog.ErrorS(err, "Failed running PreBind plugin", "plugin", pl.Name(), "pod", klog.KObj(pod))
			return framework.AsStatus(fmt.Errorf("running PreBind plugin %q: %w", pl.Name(), err))
		}
	}
	return nil
}
func (f *Framework) runPreBindPlugin(ctx context.Context, pl framework.PreBindPlugin, state *framework.CycleState,
	pod *corev1.Pod, nodeName string) *framework.Status {
	return pl.PreBind(ctx, state, pod, nodeName)
}

func (f *Framework) RunBindPlugins(ctx context.Context, state *framework.CycleState, pod *corev1.Pod, nodeName string) *framework.Status {
	var status *framework.Status
	if len(f.bindPlugins) == 0 {
		return framework.NewStatus(framework.Skip, "")
	}
	for _, bp := range f.bindPlugins {
		status = f.runBindPlugin(ctx, bp, state, pod, nodeName)
		if status != nil && status.Code() == framework.Skip {
			continue
		}
		if !status.IsSuccess() {
			err := status.AsError()
			klog.ErrorS(err, "Failed running Bind plugin", "plugin", bp.Name(), "pod", klog.KObj(pod))
			return framework.AsStatus(fmt.Errorf("running Bind plugin %q: %w", bp.Name(), err))
		}
		return status
	}
	return status
}
func (f *Framework) runBindPlugin(ctx context.Context, pl framework.BindPlugin, state *framework.CycleState,
	pod *corev1.Pod, nodeName string) *framework.Status {
	return pl.Bind(ctx, state, pod, nodeName)
}

func (f *Framework) RunPostBindPlugins(ctx context.Context, state *framework.CycleState, pod *corev1.Pod, nodeName string) {
	for _, pl := range f.postBindPlugins {
		f.runPostBindPlugin(ctx, pl, state, pod, nodeName)
	}
}
func (f *Framework) runPostBindPlugin(ctx context.Context, pl framework.PostBindPlugin, state *framework.CycleState, pod *corev1.Pod, nodeName string) {
	pl.PostBind(ctx, state, pod, nodeName)
}
