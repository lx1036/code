package manager

import (
	"flag"
	"fmt"
	"math"
	"math/rand"
	"sort"
	"sync"
	"time"

	"k8s-lx1036/k8s/kubelet/pkg/cadvisor/pkg/cache/memory"
	"k8s-lx1036/k8s/kubelet/pkg/cadvisor/pkg/container"
	"k8s-lx1036/k8s/kubelet/pkg/cadvisor/pkg/info/v1"

	"k8s.io/klog/v2"
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
	go cd.housekeeping()

	return nil
}

// 每一个容器都会启动一个 goroutine 来 collect container data
func (cd *containerData) housekeeping() {
	// Start any background goroutines - must be cleaned up in cd.handler.Cleanup().
	cd.handler.Start()
	defer cd.handler.Cleanup() // goroutine stop 之后需要 cleanup

	// Long housekeeping is either 100ms or half of the housekeeping interval.
	longHousekeeping := 100 * time.Millisecond
	if *HousekeepingInterval/2 < longHousekeeping {
		longHousekeeping = *HousekeepingInterval / 2
	}

	// Housekeep every second.
	klog.V(3).Infof("Start housekeeping for container %q\n", cd.info.Name)
	houseKeepingTimer := cd.clock.NewTimer(0 * time.Second)
	defer houseKeepingTimer.Stop()

	for {
		if !cd.housekeepingTick(houseKeepingTimer.C(), longHousekeeping) {
			return
		}
		// Stop and drain the timer so that it is safe to reset it
		if !houseKeepingTimer.Stop() {
			select {
			case <-houseKeepingTimer.C():
			default:
			}
		}

		houseKeepingTimer.Reset(cd.nextHousekeepingInterval())
	}
}

// Determine when the next housekeeping should occur.
func (cd *containerData) nextHousekeepingInterval() time.Duration {
	if cd.allowDynamicHousekeeping {
		var empty time.Time
		stats, err := cd.memoryCache.RecentStats(cd.info.Name, empty, empty, 2)
		if err != nil {
			klog.Warningf("Failed to get RecentStats(%q) while determining the next housekeeping: %v", cd.info.Name, err)
		} else if len(stats) == 2 {
			// TODO(vishnuk): Use no processes as a signal.
			// Raise the interval if usage hasn't changed in the last housekeeping.
			if stats[0].StatsEq(stats[1]) && (cd.housekeepingInterval < cd.maxHousekeepingInterval) {
				cd.housekeepingInterval *= 2
				if cd.housekeepingInterval > cd.maxHousekeepingInterval {
					cd.housekeepingInterval = cd.maxHousekeepingInterval
				}
			} else if cd.housekeepingInterval != *HousekeepingInterval {
				// Lower interval back to the baseline.
				cd.housekeepingInterval = *HousekeepingInterval
			}
		}
	}

	return jitter(cd.housekeepingInterval, 1.0)
}

// jitter returns a time.Duration between duration and duration + maxFactor * duration,
// to allow clients to avoid converging on periodic behavior.  If maxFactor is 0.0, a
// suggested default value will be chosen.
func jitter(duration time.Duration, maxFactor float64) time.Duration {
	if maxFactor <= 0.0 {
		maxFactor = 1.0
	}
	wait := duration + time.Duration(rand.Float64()*maxFactor*float64(duration))
	return wait
}

func (cd *containerData) housekeepingTick(timer <-chan time.Time, longHousekeeping time.Duration) bool {
	select {
	case <-cd.stop:
		// Stop housekeeping when signaled.
		return false
	case finishedChan := <-cd.onDemandChan:
		// notify the calling function once housekeeping has completed
		defer close(finishedChan)
	case <-timer:
	}
	start := cd.clock.Now()
	err := cd.updateStats()
	if err != nil {
		klog.Warningf("Failed to update stats for container \"%s\": %s", cd.info.Name, err)
	}
	// Log if housekeeping took too long.
	duration := cd.clock.Since(start)
	if duration >= longHousekeeping {
		klog.V(3).Infof("[%s] Housekeeping took %s", cd.info.Name, duration)
	}
	//cd.notifyOnDemand()
	cd.statsLastUpdatedTime = cd.clock.Now()
	return true
}

// INFO: 从 handler.GetStats() 获取每一个 container stats，并存入 memoryCache 对象中
func (cd *containerData) updateStats() error {
	stats, statsErr := cd.handler.GetStats()
	if statsErr != nil {
		// Ignore errors if the container is dead.
		if !cd.handler.Exists() {
			return nil
		}

		// Stats may be partially populated, push those before we return an error.
		statsErr = fmt.Errorf("%v, continuing to push stats", statsErr)
	}
	if stats == nil {
		return statsErr
	}
	var customStatsErr error

	ref, err := cd.handler.ContainerReference()
	if err != nil {
		// Ignore errors if the container is dead.
		if !cd.handler.Exists() {
			return nil
		}
		return err
	}

	cInfo := v1.ContainerInfo{
		ContainerReference: ref,
	}

	err = cd.memoryCache.AddStats(&cInfo, stats)
	if err != nil {
		return err
	}
	if statsErr != nil {
		return statsErr
	}

	return customStatsErr
}

func (cd *containerData) Stop() error {
	err := cd.memoryCache.RemoveContainer(cd.info.Name)
	if err != nil {
		return err
	}
	close(cd.stop)

	return nil
}

func (cd *containerData) GetInfo(shouldUpdateSubcontainers bool) (*containerInfo, error) {
	// Get spec and subcontainers.
	if cd.clock.Since(cd.infoLastUpdatedTime) > 5*time.Second || shouldUpdateSubcontainers {
		err := cd.updateSpec()
		if err != nil {
			return nil, err
		}
		if shouldUpdateSubcontainers {
			err = cd.updateSubcontainers()
			if err != nil {
				return nil, err
			}
		}
		cd.infoLastUpdatedTime = cd.clock.Now()
	}
	cd.lock.Lock()
	defer cd.lock.Unlock()
	cInfo := containerInfo{
		Subcontainers: cd.info.Subcontainers,
		Spec:          cd.info.Spec,
	}
	cInfo.Id = cd.info.Id
	cInfo.Name = cd.info.Name
	cInfo.Aliases = cd.info.Aliases
	cInfo.Namespace = cd.info.Namespace
	return &cInfo, nil
}

func (cd *containerData) updateSpec() error {
	spec, err := cd.handler.GetSpec()
	if err != nil {
		// Ignore errors if the container is dead.
		if !cd.handler.Exists() {
			return nil
		}
		return err
	}

	/*customMetrics, err := cd.collectorManager.GetSpec()
	if err != nil {
		return err
	}
	if len(customMetrics) > 0 {
		spec.HasCustomMetrics = true
		spec.CustomMetrics = customMetrics
	}*/

	cd.lock.Lock()
	defer cd.lock.Unlock()
	cd.info.Spec = spec

	return nil
}

func (cd *containerData) updateSubcontainers() error {
	var subcontainers v1.ContainerReferenceSlice
	subcontainers, err := cd.handler.ListContainers(container.ListSelf)
	if err != nil {
		// Ignore errors if the container is dead.
		if !cd.handler.Exists() {
			return nil
		}
		return err
	}
	sort.Sort(subcontainers)
	cd.lock.Lock()
	defer cd.lock.Unlock()
	cd.info.Subcontainers = subcontainers

	return nil
}
