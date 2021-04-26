package manager

import (
	"flag"
	"fmt"
	"math"
	"sync"
	"time"

	"k8s-lx1036/k8s/kubelet/pkg/cadvisor/pkg/cache/memory"
	"k8s-lx1036/k8s/kubelet/pkg/cadvisor/pkg/container"
	"k8s-lx1036/k8s/kubelet/pkg/cadvisor/pkg/info/v1"

	"k8s.io/utils/clock"
)

var HousekeepingInterval = flag.Duration("housekeeping_interval", 1*time.Second, "Interval between container housekeepings")

type containerInfo struct {
	v1.ContainerReference
	Subcontainers []v1.ContainerReference
	Spec          v1.ContainerSpec
}

type containerData struct {
	handler     container.ContainerHandler
	info        containerInfo
	memoryCache *memory.InMemoryCache
	lock        sync.Mutex
	//loadReader               cpuload.CpuLoadReader
	//summaryReader            *summary.StatsSummary
	loadAvg                  float64 // smoothed load average seen so far.
	housekeepingInterval     time.Duration
	maxHousekeepingInterval  time.Duration
	allowDynamicHousekeeping bool
	infoLastUpdatedTime      time.Time
	statsLastUpdatedTime     time.Time
	lastErrorTime            time.Time
	//  used to track time
	clock clock.Clock

	// Decay value used for load average smoothing. Interval length of 10 seconds is used.
	loadDecay float64

	// Whether to log the usage of this container when it is updated.
	logUsage bool

	// Tells the container to stop.
	stop chan struct{}

	// Tells the container to immediately collect stats
	onDemandChan chan chan struct{}

	// Runs custom metric collectors.
	//collectorManager collector.CollectorManager

	// nvidiaCollector updates stats for Nvidia GPUs attached to the container.
	//nvidiaCollector stats.Collector

	// perfCollector updates stats for perf_event cgroup controller.
	//perfCollector stats.Collector

	// resctrlCollector updates stats for resctrl controller.
	//resctrlCollector stats.Collector
}

func newContainerData(containerName string, memoryCache *memory.InMemoryCache, handler container.ContainerHandler,
	logUsage bool, maxHousekeepingInterval time.Duration,
	allowDynamicHousekeeping bool, clock clock.Clock) (*containerData, error) {
	if memoryCache == nil {
		return nil, fmt.Errorf("nil memory storage")
	}
	if handler == nil {
		return nil, fmt.Errorf("nil container handler")
	}
	ref, err := handler.ContainerReference()
	if err != nil {
		return nil, err
	}

	cont := &containerData{
		handler:                  handler,
		memoryCache:              memoryCache,
		housekeepingInterval:     *HousekeepingInterval,
		maxHousekeepingInterval:  maxHousekeepingInterval,
		allowDynamicHousekeeping: allowDynamicHousekeeping,
		logUsage:                 logUsage,
		loadAvg:                  -1.0, // negative value indicates uninitialized.
		stop:                     make(chan struct{}),
		//collectorManager:         collectorManager,
		onDemandChan: make(chan chan struct{}, 100),
		clock:        clock,
	}
	cont.info.ContainerReference = ref

	cont.loadDecay = math.Exp(float64(-cont.housekeepingInterval.Seconds() / 10))

	/*if *enableLoadReader {
		// Create cpu load reader.
		loadReader, err := cpuload.New()
		if err != nil {
			klog.Warningf("Could not initialize cpu load reader for %q: %s", ref.Name, err)
		} else {
			cont.loadReader = loadReader
		}
	}*/

	/*err = cont.updateSpec()
	if err != nil {
		return nil, err
	}*/
	/*cont.summaryReader, err = summary.New(cont.info.Spec)
	if err != nil {
		cont.summaryReader = nil
		klog.Infof("Failed to create summary reader for %q: %v", ref.Name, err)
	}*/

	return cont, nil
}

func (cd *containerData) Start() error {
	//go cd.housekeeping()
	return nil
}

func (cd *containerData) Stop() error {
	err := cd.memoryCache.RemoveContainer(cd.info.Name)
	if err != nil {
		return err
	}
	close(cd.stop)

	return nil
}
