package pkg

import (
	"errors"
	"fmt"
	"time"

	"k8s-lx1036/k8s/scheduler/pkg/algorithmprovider"
	schedulerapi "k8s-lx1036/k8s/scheduler/pkg/apis/config"
	"k8s-lx1036/k8s/scheduler/pkg/core"
	framework "k8s-lx1036/k8s/scheduler/pkg/framework"
	frameworkruntime "k8s-lx1036/k8s/scheduler/pkg/framework/runtime"
	internalcache "k8s-lx1036/k8s/scheduler/pkg/internal/cache"
	internalqueue "k8s-lx1036/k8s/scheduler/pkg/internal/queue"
	"k8s-lx1036/k8s/scheduler/pkg/profile"

	"k8s.io/client-go/informers"
	coreinformers "k8s.io/client-go/informers/core/v1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
)

// Configurator defines I/O, caching, and other functionality needed to
// construct a new scheduler.
type Configurator struct {
	client clientset.Interface

	recorderFactory profile.RecorderFactory

	informerFactory informers.SharedInformerFactory

	podInformer coreinformers.PodInformer

	// Close this to stop all reflectors
	StopEverything <-chan struct{}

	schedulerCache internalcache.Cache

	// Disable pod preemption or not.
	disablePreemption bool

	// Always check all predicates even if the middle of one predicate fails.
	alwaysCheckAllPredicates bool

	// percentageOfNodesToScore specifies percentage of all nodes to score in each scheduling cycle.
	percentageOfNodesToScore int32

	podInitialBackoffSeconds int64

	podMaxBackoffSeconds int64

	profiles          []schedulerapi.KubeSchedulerProfile
	registry          frameworkruntime.Registry
	nodeInfoSnapshot  *internalcache.Snapshot
	extenders         []schedulerapi.Extender
	frameworkCapturer FrameworkCapturer
}

func (c *Configurator) createFromProvider(providerName string) (*Scheduler, error) {
	klog.V(2).Infof("Creating scheduler from algorithm provider '%v'", providerName)
	r := algorithmprovider.NewRegistry()
	defaultPlugins, exist := r[providerName]
	if !exist {
		return nil, fmt.Errorf("algorithm provider %q is not registered", providerName)
	}

	// 把config.yaml里的profiles与默认的algorithmprovider/registry，merge下
	// 合并config.yaml profiles定义的plugins，和kube-scheduler在各个hooks定义的default plugins
	// INFO: 这里逻辑可以 disable kube-scheduler 默认的 plugin，比如 disable "NodeResourcesLeastAllocated" plugin, 在 scheduler-config.yaml 对象里配置
	for i := range c.profiles {
		prof := &c.profiles[i]
		plugins := &schedulerapi.Plugins{}
		plugins.Append(defaultPlugins)
		plugins.Apply(prof.Plugins)
		prof.Plugins = plugins
	}

	return c.create()
}

// create a scheduler from a set of registered plugins.
func (c *Configurator) create() (*Scheduler, error) {
	// The nominator will be passed all the way to framework instantiation.
	nominator := internalqueue.NewPodNominator()
	profiles, err := profile.NewMap(c.profiles, c.buildFramework, c.recorderFactory,
		frameworkruntime.WithPodNominator(nominator))
	if err != nil {
		return nil, fmt.Errorf("initializing profiles: %v", err)
	}
	if len(profiles) == 0 {
		return nil, errors.New("at least one profile is required")
	}
	// Profiles are required to have equivalent queue sort plugins.
	lessFn := profiles[c.profiles[0].SchedulerName].Framework.QueueSortFunc()
	podQueue := internalqueue.NewSchedulingQueue(
		lessFn,
		internalqueue.WithPodInitialBackoffDuration(time.Duration(c.podInitialBackoffSeconds)*time.Second),
		internalqueue.WithPodMaxBackoffDuration(time.Duration(c.podMaxBackoffSeconds)*time.Second),
		internalqueue.WithPodNominator(nominator),
	)

	algo := core.NewGenericScheduler(
		c.schedulerCache,
		c.nodeInfoSnapshot,
		c.informerFactory.Core().V1().PersistentVolumeClaims().Lister(),
		c.disablePreemption,
		c.percentageOfNodesToScore,
	)

	return &Scheduler{
		SchedulerCache:  c.schedulerCache,
		Algorithm:       algo,
		Profiles:        profiles,
		NextPod:         internalqueue.MakeNextPodFunc(podQueue),
		Error:           MakeDefaultErrorFunc(c.client, c.informerFactory.Core().V1().Pods().Lister(), podQueue, c.schedulerCache),
		StopEverything:  c.StopEverything,
		SchedulingQueue: podQueue,
	}, nil
}

func (c *Configurator) buildFramework(p schedulerapi.KubeSchedulerProfile, opts ...frameworkruntime.Option) (framework.Framework, error) {
	if c.frameworkCapturer != nil {
		c.frameworkCapturer(p)
	}
	opts = append([]frameworkruntime.Option{
		frameworkruntime.WithClientSet(c.client),
		frameworkruntime.WithInformerFactory(c.informerFactory),
		frameworkruntime.WithSnapshotSharedLister(c.nodeInfoSnapshot),
		frameworkruntime.WithRunAllFilters(c.alwaysCheckAllPredicates),
	}, opts...)

	return frameworkruntime.NewFramework(
		c.registry,
		p.Plugins,
		p.PluginConfig,
		opts...,
	)
}
