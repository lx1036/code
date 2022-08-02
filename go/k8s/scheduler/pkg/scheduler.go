package pkg

import (
	"context"
	"errors"
	"fmt"
	"k8s.io/client-go/dynamic/dynamicinformer"
	corelisters "k8s.io/client-go/listers/core/v1"
	restclient "k8s.io/client-go/rest"
	"time"

	"k8s-lx1036/k8s/scheduler/pkg/apis/config"
	"k8s-lx1036/k8s/scheduler/pkg/core"
	"k8s-lx1036/k8s/scheduler/pkg/framework"
	frameworkplugins "k8s-lx1036/k8s/scheduler/pkg/framework/plugins"
	frameworkruntime "k8s-lx1036/k8s/scheduler/pkg/framework/runtime"
	internalcache "k8s-lx1036/k8s/scheduler/pkg/internal/cache"
	internalqueue "k8s-lx1036/k8s/scheduler/pkg/internal/queue"
	"k8s-lx1036/k8s/scheduler/pkg/metrics"
	"k8s-lx1036/k8s/scheduler/pkg/util"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	coreinformers "k8s.io/client-go/informers/core/v1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
	podutil "k8s.io/kubernetes/pkg/api/v1/pod"
	"k8s.io/kubernetes/pkg/apis/core/validation"
)

const (
	pluginMetricsSamplePercent = 10

	// Duration the scheduler will wait before expiring an assumed pod.
	// See issue #106361 for more details about this parameter and its value.
	durationToExpireAssumedPod = 15 * time.Minute
)

func init() {
	metrics.Register()
}

// FrameworkCapturer is used for registering a notify function in building framework.
type FrameworkCapturer func(config.KubeSchedulerProfile)

type Scheduler struct {
	Frameworks frameworkruntime.Frameworks

	SchedulerCache *internalcache.Cache
	PriorityQueue  *internalqueue.PriorityQueue

	SchedulePod func(ctx context.Context, fwk framework.Framework,
		state *framework.CycleState, pod *v1.Pod) (ScheduleResult, error)

	// INFO: 类似 iptables，把各个hooks串起来
	Algorithm core.ScheduleAlgorithm

	// NextPod should be a function that blocks until the next pod
	// is available. We don't use a channel for this, because scheduling
	// a pod may take some amount of time and we don't want pods to get
	// stale while they sit in a channel.
	NextPod func() *framework.QueuedPodInfo

	// Error is called if there is an error. It is passed the pod in
	// question, and the error
	Error func(*framework.QueuedPodInfo, error)

	// Close this to shut down the scheduler.
	StopEverything <-chan struct{}

	scheduledPodsHasSynced func() bool

	client clientset.Interface

	profiles                 []config.KubeSchedulerProfile
	podInitialBackoffSeconds int64
	podMaxBackoffSeconds     int64
	//recorderFactory profile.RecorderFactory
	informerFactory informers.SharedInformerFactory
	podInformer     coreinformers.PodInformer

	// Disable pod preemption or not.
	disablePreemption bool
	// Always check all predicates even if the middle of one predicate fails.
	alwaysCheckAllPredicates bool
	// percentageOfNodesToScore specifies percentage of all nodes to score in each scheduling cycle.
	percentageOfNodesToScore int32
	registry                 frameworkruntime.Registry
	nodeInfoSnapshot         *internalcache.Snapshot
	frameworkCapturer        FrameworkCapturer
}

type schedulerOptions struct {
	kubeConfig               *restclient.Config
	schedulerAlgorithmSource config.SchedulerAlgorithmSource
	percentageOfNodesToScore int32
	podInitialBackoffSeconds int64
	podMaxBackoffSeconds     int64
	// Contains out-of-tree plugins to be merged with the in-tree registry.
	frameworkOutOfTreeRegistry frameworkruntime.Registry
	profiles                   []config.KubeSchedulerProfile
	extenders                  []config.Extender
	frameworkCapturer          FrameworkCapturer
	parallelism                int32 // sets the parallelism for all scheduler algorithms. Default is 16.
}

// Option configures a Scheduler
type Option func(*schedulerOptions)

var defaultSchedulerOptions = schedulerOptions{
	percentageOfNodesToScore: config.DefaultPercentageOfNodesToScore,
	podInitialBackoffSeconds: int64(internalqueue.DefaultPodInitialBackoffDuration.Seconds()),
	podMaxBackoffSeconds:     int64(internalqueue.DefaultPodMaxBackoffDuration.Seconds()),
}

func MakeDefaultErrorFunc(client clientset.Interface, podLister corelisters.PodLister,
	podQueue *internalqueue.PriorityQueue, schedulerCache *internalcache.Cache) func(*framework.QueuedPodInfo, error) {
	return func(podInfo *framework.QueuedPodInfo, err error) {

	}
}

func New(
	client clientset.Interface,
	informerFactory informers.SharedInformerFactory,
	dynInformerFactory dynamicinformer.DynamicSharedInformerFactory,
	recorderFactory frameworkruntime.RecorderFactory,
	stopCh <-chan struct{}, opts ...Option) (*Scheduler, error) {
	// INFO: scheduler提供了一套扩展机制 scheduler-framework，用来可以合并 out-of-tree registry plugins，其实也比较简单!!!
	//  (1)merge registry plugin
	stopEverything := stopCh
	if stopEverything == nil {
		stopEverything = wait.NeverStop
	}
	options := defaultSchedulerOptions
	for _, opt := range opts {
		opt(&options)
	}
	registry := frameworkplugins.NewInTreeRegistry()
	if err := registry.Merge(options.frameworkOutOfTreeRegistry); err != nil {
		return nil, err
	}

	// (2)new frameworks
	clusterEventMap := make(map[framework.ClusterEvent]sets.String)
	podLister := informerFactory.Core().V1().Pods().Lister()
	snapshot := internalcache.NewEmptySnapshot()
	nominator := internalqueue.NewPodNominator(podLister)
	frameworks, err := frameworkruntime.NewFrameworks(options.profiles, registry, recorderFactory,
		//frameworkruntime.WithComponentConfigVersion(options.componentConfigVersion),
		frameworkruntime.WithClientSet(client),
		frameworkruntime.WithKubeConfig(options.kubeConfig),
		frameworkruntime.WithInformerFactory(informerFactory),
		frameworkruntime.WithSnapshotSharedLister(snapshot),
		frameworkruntime.WithPodNominator(nominator),
		frameworkruntime.WithCaptureProfile(frameworkruntime.CaptureProfile(options.frameworkCapturer)),
		frameworkruntime.WithClusterEventMap(clusterEventMap),
		frameworkruntime.WithParallelism(int(options.parallelism)),
	)
	if err != nil {
		return nil, fmt.Errorf("initializing profiles: %v", err)
	}
	if len(frameworks) == 0 {
		return nil, errors.New("at least one profile is required")
	}

	// (3)new scheduler
	podQueue := internalqueue.NewPriorityQueue(
		frameworks[options.profiles[0].SchedulerName].QueueSortFunc(),
		informerFactory,
		internalqueue.WithPodInitialBackoffDuration(time.Duration(options.podInitialBackoffSeconds)*time.Second),
		internalqueue.WithPodMaxBackoffDuration(time.Duration(options.podMaxBackoffSeconds)*time.Second),
		internalqueue.WithPodNominator(nominator),
		internalqueue.WithClusterEventMap(clusterEventMap),
	)
	schedulerCache := internalcache.New(durationToExpireAssumedPod, stopEverything)
	sched := newScheduler(
		schedulerCache,
		internalqueue.MakeNextPodFunc(podQueue),
		MakeDefaultErrorFunc(client, podLister, podQueue, schedulerCache),
		stopEverything,
		podQueue,
		frameworks,
		client,
		snapshot,
		options.percentageOfNodesToScore,
	)

	addAllEventHandlers(sched, informerFactory, dynInformerFactory, unionedGVKs(clusterEventMap))

	return sched, nil
}

// for test case
func newScheduler(
	cache *internalcache.Cache,
	nextPod func() *framework.QueuedPodInfo,
	Error func(*framework.QueuedPodInfo, error),
	stopEverything <-chan struct{},
	podQueue *internalqueue.PriorityQueue,
	frameworks frameworkruntime.Frameworks,
	client clientset.Interface,
	nodeInfoSnapshot *internalcache.Snapshot,
	percentageOfNodesToScore int32) *Scheduler {
	sched := Scheduler{
		SchedulerCache: cache,
		PriorityQueue:  podQueue,

		NextPod:                  nextPod,
		Error:                    Error,
		StopEverything:           stopEverything,
		Frameworks:               frameworks,
		client:                   client,
		nodeInfoSnapshot:         nodeInfoSnapshot,
		percentageOfNodesToScore: percentageOfNodesToScore,
	}
	sched.SchedulePod = sched.schedulePod
	return &sched
}

func (scheduler *Scheduler) Run(ctx context.Context) {
	scheduler.PriorityQueue.Run()
	wait.UntilWithContext(ctx, scheduler.scheduleOne, 0)
	scheduler.PriorityQueue.Close()
}

// recordSchedulingFailure records an event for the pod that indicates the
// pod has failed to schedule. Also, update the pod condition and nominated node name if set.
func (scheduler *Scheduler) recordSchedulingFailure(prof *frameworkruntime.Framework, podInfo *framework.QueuedPodInfo,
	err error, reason string, nominatedNode string) {
	scheduler.Error(podInfo, err)

	// Update the scheduling queue with the nominated pod information. Without
	// this, there would be a race condition between the next scheduling cycle
	// and the time the scheduler receives a Pod Update for the nominated pod.
	// Here we check for nil only for tests.
	if scheduler.PriorityQueue != nil {
		scheduler.PriorityQueue.AddNominatedPod(podInfo.Pod, nominatedNode)
	}

	pod := podInfo.Pod
	//msg := truncateMessage(err.Error())
	//prof.Recorder.Eventf(pod, nil, v1.EventTypeWarning, "FailedScheduling", "Scheduling", msg)
	if err := updatePod(scheduler.client, pod, &v1.PodCondition{
		Type:    v1.PodScheduled,
		Status:  v1.ConditionFalse,
		Reason:  reason,
		Message: err.Error(),
	}, nominatedNode); err != nil {
		klog.Errorf("Error updating pod %s/%s: %v", pod.Namespace, pod.Name, err)
	}
}

// truncateMessage truncates a message if it hits the NoteLengthLimit.
func truncateMessage(message string) string {
	max := validation.NoteLengthLimit
	if len(message) <= max {
		return message
	}
	suffix := " ..."
	return message[:max-len(suffix)] + suffix
}

func updatePod(client clientset.Interface, pod *v1.Pod, condition *v1.PodCondition, nominatedNode string) error {
	klog.V(3).Infof("Updating pod condition for %s/%s to (%s==%s, Reason=%s)",
		pod.Namespace, pod.Name, condition.Type, condition.Status, condition.Reason)
	podCopy := pod.DeepCopy()
	// NominatedNodeName is updated only if we are trying to set it, and the value is
	// different from the existing one.
	if !podutil.UpdatePodCondition(&podCopy.Status, condition) &&
		(len(nominatedNode) == 0 || pod.Status.NominatedNodeName == nominatedNode) {
		return nil
	}
	if nominatedNode != "" {
		podCopy.Status.NominatedNodeName = nominatedNode
	}

	return util.PatchPod(client, pod, podCopy)
}

func WithProfiles(p ...config.KubeSchedulerProfile) Option {
	return func(o *schedulerOptions) {
		o.profiles = p
	}
}

func WithPercentageOfNodesToScore(percentageOfNodesToScore int32) Option {
	return func(o *schedulerOptions) {
		o.percentageOfNodesToScore = percentageOfNodesToScore
	}
}

func WithFrameworkOutOfTreeRegistry(registry frameworkruntime.Registry) Option {
	return func(o *schedulerOptions) {
		o.frameworkOutOfTreeRegistry = registry
	}
}

func WithPodInitialBackoffSeconds(podInitialBackoffSeconds int64) Option {
	return func(o *schedulerOptions) {
		o.podInitialBackoffSeconds = podInitialBackoffSeconds
	}
}

func WithPodMaxBackoffSeconds(podMaxBackoffSeconds int64) Option {
	return func(o *schedulerOptions) {
		o.podMaxBackoffSeconds = podMaxBackoffSeconds
	}
}

// WithParallelism sets the parallelism for all scheduler algorithms. Default is 16.
func WithParallelism(threads int32) Option {
	return func(o *schedulerOptions) {
		o.parallelism = threads
	}
}

func unionedGVKs(m map[framework.ClusterEvent]sets.String) map[framework.GVK]framework.ActionType {
	gvkMap := make(map[framework.GVK]framework.ActionType)
	for evt := range m {
		if _, ok := gvkMap[evt.Resource]; ok {
			gvkMap[evt.Resource] |= evt.ActionType
		} else {
			gvkMap[evt.Resource] = evt.ActionType
		}
	}
	return gvkMap
}
