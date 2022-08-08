package pkg

import (
	"context"
	"errors"
	"fmt"
	configv1 "k8s-lx1036/k8s/scheduler/pkg/apis/config/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic/dynamicinformer"
	corelisters "k8s.io/client-go/listers/core/v1"
	restclient "k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	"time"

	"k8s-lx1036/k8s/scheduler/pkg/framework"
	frameworkplugins "k8s-lx1036/k8s/scheduler/pkg/framework/plugins"
	frameworkruntime "k8s-lx1036/k8s/scheduler/pkg/framework/runtime"
	internalcache "k8s-lx1036/k8s/scheduler/pkg/internal/cache"
	internalqueue "k8s-lx1036/k8s/scheduler/pkg/internal/queue"
	"k8s-lx1036/k8s/scheduler/pkg/metrics"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	coreinformers "k8s.io/client-go/informers/core/v1"
	clientset "k8s.io/client-go/kubernetes"
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
type FrameworkCapturer func(configv1.KubeSchedulerProfile)

type Scheduler struct {
	Frameworks frameworkruntime.Frameworks

	SchedulerCache *internalcache.Cache
	PriorityQueue  *internalqueue.PriorityQueue

	SchedulePod func(ctx context.Context, fwk *frameworkruntime.Framework,
		state *framework.CycleState, pod *v1.Pod) (ScheduleResult, error)

	NextPod func() *framework.QueuedPodInfo

	Error func(*framework.QueuedPodInfo, error)

	// Close this to shut down the scheduler.
	StopEverything <-chan struct{}

	scheduledPodsHasSynced func() bool

	client clientset.Interface

	profiles                 []configv1.KubeSchedulerProfile
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

	nextStartNodeIndex int
}

type schedulerOptions struct {
	kubeConfig               *restclient.Config
	schedulerAlgorithmSource configv1.SchedulerAlgorithmSource
	percentageOfNodesToScore int32
	podInitialBackoffSeconds int64
	podMaxBackoffSeconds     int64
	// Contains out-of-tree plugins to be merged with the in-tree registry.
	frameworkOutOfTreeRegistry frameworkruntime.Registry
	profiles                   []configv1.KubeSchedulerProfile
	extenders                  []configv1.Extender
	frameworkCapturer          FrameworkCapturer
	parallelism                int32 // sets the parallelism for all scheduler algorithms. Default is 16.
}

// Option configures a Scheduler
type Option func(*schedulerOptions)

var defaultSchedulerOptions = schedulerOptions{
	percentageOfNodesToScore: configv1.DefaultPercentageOfNodesToScore,
	podInitialBackoffSeconds: int64(internalqueue.DefaultPodInitialBackoffDuration.Seconds()),
	podMaxBackoffSeconds:     int64(internalqueue.DefaultPodMaxBackoffDuration.Seconds()),
}

func MakeDefaultErrorFunc(client clientset.Interface, podLister corelisters.PodLister,
	podQueue *internalqueue.PriorityQueue, schedulerCache *internalcache.Cache) func(*framework.QueuedPodInfo, error) {
	return func(podInfo *framework.QueuedPodInfo, err error) {
		pod := podInfo.Pod
		if err == ErrNoNodesAvailable {
			klog.V(2).InfoS("Unable to schedule pod; no nodes are registered to the cluster; waiting", "pod", klog.KObj(pod))
		} else if fitError, ok := err.(*framework.FitError); ok {
			// Inject UnschedulablePlugins to PodInfo, which will be used later for moving Pods between queues efficiently.
			podInfo.UnschedulablePlugins = fitError.Diagnosis.UnschedulablePlugins
			klog.V(2).InfoS("Unable to schedule pod; no fit; waiting", "pod", klog.KObj(pod), "err", err)
		} else if apierrors.IsNotFound(err) {
			klog.V(2).InfoS("Unable to schedule pod, possibly due to node not found; waiting", "pod", klog.KObj(pod), "err", err)
			if errStatus, ok := err.(apierrors.APIStatus); ok && errStatus.Status().Details.Kind == "node" {
				nodeName := errStatus.Status().Details.Name
				// when node is not found, We do not remove the node right away. Trying again to get
				// the node and if the node is still not found, then remove it from the scheduler cache.
				_, err := client.CoreV1().Nodes().Get(context.TODO(), nodeName, metav1.GetOptions{})
				if err != nil && apierrors.IsNotFound(err) {
					node := v1.Node{ObjectMeta: metav1.ObjectMeta{Name: nodeName}}
					if err := schedulerCache.RemoveNode(&node); err != nil {
						klog.V(4).InfoS("Node is not found; failed to remove it from the cache", "node", node.Name)
					}
				}
			}
		} else {
			klog.ErrorS(err, "Error scheduling pod; retrying", "pod", klog.KObj(pod))
		}

		// Check if the Pod exists in informer cache.
		cachedPod, err := podLister.Pods(pod.Namespace).Get(pod.Name)
		if err != nil {
			klog.InfoS("Pod doesn't exist in informer cache", "pod", klog.KObj(pod), "err", err)
			return
		}
		if len(cachedPod.Spec.NodeName) != 0 {
			klog.InfoS("Pod has been assigned to node. Abort adding it back to queue.", "pod", klog.KObj(pod), "node", cachedPod.Spec.NodeName)
			return
		}
		// As <cachedPod> is from SharedInformer, we need to do a DeepCopy() here.
		podInfo.PodInfo = framework.NewPodInfo(cachedPod.DeepCopy())
		if err := podQueue.AddUnschedulableIfNotPresent(podInfo, podQueue.SchedulingCycle()); err != nil {
			klog.ErrorS(err, "Error occurred")
		}
	}
}
func WithProfiles(p ...configv1.KubeSchedulerProfile) Option {
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

func New(
	client clientset.Interface,
	informerFactory informers.SharedInformerFactory,
	dynInformerFactory dynamicinformer.DynamicSharedInformerFactory,
	recorderFactory frameworkruntime.RecorderFactory,
	stopCh <-chan struct{}, opts ...Option) (*Scheduler, error) {
	// INFO: (1)merge registry plugin, scheduler提供了一套扩展机制 scheduler-framework，用来可以合并 out-of-tree registry plugins，其实也比较简单!!!
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
	podNominator := internalqueue.NewPodNominator(podLister)
	frameworks, err := frameworkruntime.NewFrameworks(options.profiles, registry, recorderFactory,
		//frameworkruntime.WithComponentConfigVersion(options.componentConfigVersion),
		frameworkruntime.WithClientSet(client),
		frameworkruntime.WithKubeConfig(options.kubeConfig),
		frameworkruntime.WithInformerFactory(informerFactory),
		frameworkruntime.WithSnapshotSharedLister(snapshot),
		frameworkruntime.WithPodNominator(podNominator),
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
		internalqueue.WithPodNominator(podNominator),
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
	sched.SchedulePod = sched.schedulePod // for test case
	return &sched
}

func (scheduler *Scheduler) Run(ctx context.Context) {
	scheduler.PriorityQueue.Run()
	wait.UntilWithContext(ctx, scheduler.scheduleOne, 0) // block
	scheduler.PriorityQueue.Close()
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
