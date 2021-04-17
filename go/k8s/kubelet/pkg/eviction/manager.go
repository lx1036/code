package eviction

import (
	"sort"
	"sync"
	"time"

	"k8s-lx1036/k8s/kubelet/pkg/lifecycle"
	"k8s-lx1036/k8s/kubelet/pkg/server/stats"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/clock"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/kubelet/util/format"
)

// Manager evaluates when an eviction threshold for node stability has been met on the node.
type Manager interface {
	// Start starts the control loop to monitor eviction thresholds at specified interval.
	Start(diskInfoProvider DiskInfoProvider, podFunc ActivePodsFunc, podCleanedUpFunc PodCleanedUpFunc, monitoringInterval time.Duration)

	// IsUnderMemoryPressure returns true if the node is under memory pressure.
	IsUnderMemoryPressure() bool

	// IsUnderDiskPressure returns true if the node is under disk pressure.
	IsUnderDiskPressure() bool

	// IsUnderPIDPressure returns true if the node is under PID pressure.
	IsUnderPIDPressure() bool
}

type managerImpl struct {
	//  used to track time
	clock clock.Clock
	// config is how the manager is configured
	config Config
	// the function to invoke to kill a pod
	killPodFunc KillPodFunc
	// the function to get the mirror pod by a given statid pod
	mirrorPodFunc MirrorPodFunc
	// the interface that knows how to do image gc
	imageGC ImageGC
	// the interface that knows how to do container gc
	containerGC ContainerGC
	// protects access to internal state
	sync.RWMutex
	// node conditions are the set of conditions present
	nodeConditions []v1.NodeConditionType
	// captures when a node condition was last observed based on a threshold being met
	nodeConditionsLastObservedAt nodeConditionsObservedAt
	// nodeRef is a reference to the node
	nodeRef *v1.ObjectReference
	// used to record events about the node
	recorder record.EventRecorder
	// used to measure usage stats on system
	summaryProvider stats.SummaryProvider
	// records when a threshold was first observed
	thresholdsFirstObservedAt thresholdsObservedAt
	// records the set of thresholds that have been met (including graceperiod) but not yet resolved
	thresholdsMet []Threshold
	// signalToRankFunc maps a resource to ranking function for that resource.
	signalToRankFunc map[Signal]rankFunc
	// signalToNodeReclaimFuncs maps a resource to an ordered list of functions that know how to reclaim that resource.
	signalToNodeReclaimFuncs map[Signal]nodeReclaimFuncs
	// last observations from synchronize
	lastObservations signalObservations
	// dedicatedImageFs indicates if imagefs is on a separate device from the rootfs
	dedicatedImageFs *bool
	// thresholdNotifiers is a list of memory threshold notifiers which each notify for a memory eviction threshold
	thresholdNotifiers []ThresholdNotifier
	// thresholdsLastUpdated is the last time the thresholdNotifiers were updated.
	thresholdsLastUpdated time.Time
	// etcHostsPath is a function that will get the etc-hosts file's path for a pod given its UID
	etcHostsPath func(podUID types.UID) string
}

func (m *managerImpl) IsUnderMemoryPressure() bool {
	m.RLock()
	defer m.RUnlock()
	return hasNodeCondition(m.nodeConditions, v1.NodeMemoryPressure)
}

func (m *managerImpl) IsUnderDiskPressure() bool {
	m.RLock()
	defer m.RUnlock()
	return hasNodeCondition(m.nodeConditions, v1.NodeDiskPressure)
}

func (m *managerImpl) IsUnderPIDPressure() bool {
	m.RLock()
	defer m.RUnlock()
	return hasNodeCondition(m.nodeConditions, v1.NodePIDPressure)
}

func (m *managerImpl) Admit(attrs *lifecycle.PodAdmitAttributes) lifecycle.PodAdmitResult {
	panic("implement me")
}

// Start starts the control loop to observe and response to low compute resources.
func (m *managerImpl) Start(diskInfoProvider DiskInfoProvider, podFunc ActivePodsFunc, podCleanedUpFunc PodCleanedUpFunc, monitoringInterval time.Duration) {
	// start the eviction manager monitoring
	go wait.Until(func() {
		if evictedPods := m.synchronize(diskInfoProvider, podFunc); evictedPods != nil {
			klog.Infof("eviction manager: pods %s evicted, waiting for pod to be cleaned up", format.Pods(evictedPods))
			m.waitForPodsCleanup(podCleanedUpFunc, evictedPods)
		}
	}, monitoringInterval, wait.NeverStop)
}

// rankMemoryPressure orders the input pods for eviction in response to memory pressure.
// It ranks by whether or not the pod's usage exceeds its requests, then by priority, and
// finally by memory usage above requests.
func rankMemoryPressure(pods []*v1.Pod, stats statsFunc) {
	orderedBy(exceedMemoryRequests(stats), priority, memory(stats)).Sort(pods)
}

// rankPIDPressure orders the input pods by priority in response to PID pressure.
func rankPIDPressure(pods []*v1.Pod, stats statsFunc) {
	orderedBy(priority, process(stats)).Sort(pods)
}

// rankDiskPressureFunc returns a rankFunc that measures the specified fs stats.
func rankDiskPressureFunc(fsStatsToMeasure []fsStatsType, diskResource v1.ResourceName) rankFunc {
	return func(pods []*v1.Pod, stats statsFunc) {
		orderedBy(exceedDiskRequests(stats, fsStatsToMeasure, diskResource), priority, disk(stats, fsStatsToMeasure, diskResource)).Sort(pods)
	}
}

// buildSignalToRankFunc returns ranking functions associated with resources
func buildSignalToRankFunc(withImageFs bool) map[Signal]rankFunc {
	signalToRankFunc := map[Signal]rankFunc{
		SignalMemoryAvailable:            rankMemoryPressure,
		SignalAllocatableMemoryAvailable: rankMemoryPressure,
		SignalPIDAvailable:               rankPIDPressure,
	}
	// usage of an imagefs is optional
	if withImageFs {
		// with an imagefs, nodefs pod rank func for eviction only includes logs and local volumes
		signalToRankFunc[SignalNodeFsAvailable] = rankDiskPressureFunc([]fsStatsType{fsStatsLogs, fsStatsLocalVolumeSource}, v1.ResourceEphemeralStorage)
		signalToRankFunc[SignalNodeFsInodesFree] = rankDiskPressureFunc([]fsStatsType{fsStatsLogs, fsStatsLocalVolumeSource}, resourceInodes)
		// with an imagefs, imagefs pod rank func for eviction only includes rootfs
		signalToRankFunc[SignalImageFsAvailable] = rankDiskPressureFunc([]fsStatsType{fsStatsRoot}, v1.ResourceEphemeralStorage)
		signalToRankFunc[SignalImageFsInodesFree] = rankDiskPressureFunc([]fsStatsType{fsStatsRoot}, resourceInodes)
	} else {
		// without an imagefs, nodefs pod rank func for eviction looks at all fs stats.
		// since imagefs and nodefs share a common device, they share common ranking functions.
		signalToRankFunc[SignalNodeFsAvailable] = rankDiskPressureFunc([]fsStatsType{fsStatsRoot, fsStatsLogs, fsStatsLocalVolumeSource}, v1.ResourceEphemeralStorage)
		signalToRankFunc[SignalNodeFsInodesFree] = rankDiskPressureFunc([]fsStatsType{fsStatsRoot, fsStatsLogs, fsStatsLocalVolumeSource}, resourceInodes)
		signalToRankFunc[SignalImageFsAvailable] = rankDiskPressureFunc([]fsStatsType{fsStatsRoot, fsStatsLogs, fsStatsLocalVolumeSource}, v1.ResourceEphemeralStorage)
		signalToRankFunc[SignalImageFsInodesFree] = rankDiskPressureFunc([]fsStatsType{fsStatsRoot, fsStatsLogs, fsStatsLocalVolumeSource}, resourceInodes)
	}
	return signalToRankFunc
}

// synchronize is the main control loop that enforces eviction thresholds.
// Returns the pod that was killed, or nil if no pod was killed.
func (m *managerImpl) synchronize(diskInfoProvider DiskInfoProvider, podFunc ActivePodsFunc) []*v1.Pod {
	// if we have nothing to do, just return
	thresholds := m.config.Thresholds
	if len(thresholds) == 0 {
		return nil
	}

	if m.dedicatedImageFs == nil {
		hasImageFs, ok := diskInfoProvider.HasDedicatedImageFs()
		if ok != nil {
			return nil
		}
		m.dedicatedImageFs = &hasImageFs
		// 给每一个signal安装对应的rank func
		m.signalToRankFunc = buildSignalToRankFunc(hasImageFs)
		m.signalToNodeReclaimFuncs = buildSignalToNodeReclaimFuncs(m.imageGC, m.containerGC, hasImageFs)
	}

	summary, err := m.summaryProvider.Get(true)
	if err != nil {
		klog.Errorf("eviction manager: failed to get summary stats: %v", err)
		return nil
	}

	// make observations and get a function to derive pod usage stats relative to those observations.
	observations, statsFunc := makeSignalObservations(summary)
	// determine the set of thresholds met independent of grace period
	thresholds = thresholdsMet(thresholds, observations, false)
	// determine the set of thresholds previously met that have not yet satisfied the associated min-reclaim
	if len(m.thresholdsMet) > 0 {
		thresholdsNotYetResolved := thresholdsMet(m.thresholdsMet, observations, true)
		thresholds = mergeThresholds(thresholds, thresholdsNotYetResolved)
	}

	// the set of node conditions that are triggered by currently observed thresholds
	nodeConditions := nodeConditions(thresholds)
	if len(nodeConditions) > 0 {
		klog.Infof("eviction manager: node conditions - observed: %v", nodeConditions)
	}

	// update internal state
	m.Lock()
	m.nodeConditions = nodeConditions
	m.thresholdsFirstObservedAt = thresholdsFirstObservedAt
	m.nodeConditionsLastObservedAt = nodeConditionsLastObservedAt
	m.thresholdsMet = thresholds
	// determine the set of thresholds whose stats have been updated since the last sync
	thresholds = thresholdsUpdatedStats(thresholds, observations, m.lastObservations)
	m.lastObservations = observations
	m.Unlock()

	if len(thresholds) == 0 {
		klog.Infof("eviction manager: no resources are starved")
		return nil
	}

	// rank the thresholds by eviction priority
	sort.Sort(byEvictionPriority(thresholds))
	thresholdToReclaim, resourceToReclaim, foundAny := getReclaimableThreshold(thresholds)
	if !foundAny {
		return nil
	}

	klog.Warningf("eviction manager: attempting to reclaim %v", resourceToReclaim)
	// record an event about the resources we are now attempting to reclaim via eviction
	m.recorder.Eventf(m.nodeRef, v1.EventTypeWarning, "EvictionThresholdMet", "Attempting to reclaim %s", resourceToReclaim)

	// check if there are node-level resources we can reclaim to reduce pressure before evicting end-user pods.
	// 1. 回收资源(memory/disk)，使得signal达到threshold以下
	if m.reclaimNodeLevelResources(thresholdToReclaim.Signal, resourceToReclaim) {
		klog.Infof("eviction manager: able to reduce %v pressure without evicting pods.", resourceToReclaim)
		return nil
	}

	// 2. 回收资源后signal还是不能达到threshold之下，只能通过杀一些优先级低的pod来回收了
	klog.Infof("eviction manager: must evict pod(s) to reclaim %v", resourceToReclaim)
	// rank the pods for eviction
	// 这里获取每一个signal的排序函数
	rank, ok := m.signalToRankFunc[thresholdToReclaim.Signal]
	if !ok {
		klog.Errorf("eviction manager: no ranking function for signal %s", thresholdToReclaim.Signal)
		return nil
	}
	activePods := podFunc()
	// the only candidates viable for eviction are those pods that had anything running.
	if len(activePods) == 0 {
		klog.Errorf("eviction manager: eviction thresholds have been met, but no pods are active to evict")
		return nil
	}

	// rank the running pods for eviction for the specified resource
	rank(activePods, statsFunc)
	klog.Infof("eviction manager: pods ranked for eviction: %s", format.Pods(activePods))

	// we kill at most a single pod during each eviction interval
	// 每次tick内最多只杀一个pod来回收资源
	for _, pod := range activePods {
		gracePeriodOverride := int64(0)
		if !isHardEvictionThreshold(thresholdToReclaim) {
			gracePeriodOverride = m.config.MaxPodGracePeriodSeconds
		}
		message, annotations := evictionMessage(resourceToReclaim, pod, statsFunc)
		if m.evictPod(pod, gracePeriodOverride, message, annotations) {
			return []*v1.Pod{pod}
		}
	}

	klog.Infof("eviction manager: unable to evict any pods from the node")
	return nil
}

// reclaimNodeLevelResources attempts to reclaim node level resources.
// returns true if thresholds were satisfied and no pod eviction is required.
// 回收资源，memory/disk fs/inode fs/image fs
func (m *managerImpl) reclaimNodeLevelResources(signalToReclaim Signal, resourceToReclaim v1.ResourceName) bool {
	reclaimFuncs := m.signalToNodeReclaimFuncs[signalToReclaim]
	for _, reclaimFunc := range reclaimFuncs {
		// attempt to reclaim the pressured resource.
		if err := reclaimFunc(); err != nil {
			klog.Warningf("eviction manager: unexpected error when attempting to reduce %v pressure: %v", resourceToReclaim, err)
		}
	}

	// 资源回收之后，再去检查是否还有signal达到了threshold
	if len(reclaimFuncs) > 0 {
		// 获取资源回收之后，当前stats数据
		summary, err := m.summaryProvider.Get(true)
		if err != nil {
			klog.Errorf("eviction manager: failed to get summary stats after resource reclaim: %v", err)
			return false
		}

		// make observations and get a function to derive pod usage stats relative to those observations.
		observations, _ := makeSignalObservations(summary)
		thresholds := thresholdsMet(m.config.Thresholds, observations, false)
		if len(thresholds) == 0 {
			return true
		}
	}

	return false
}

func NewManager(
	summaryProvider stats.SummaryProvider,
	config Config,
	killPodFunc KillPodFunc,
	mirrorPodFunc MirrorPodFunc,
	imageGC ImageGC,
	containerGC ContainerGC,
	recorder record.EventRecorder,
	nodeRef *v1.ObjectReference,
	clock clock.Clock,
	etcHostsPath func(types.UID) string) (Manager, lifecycle.PodAdmitHandler) {

	manager := &managerImpl{
		clock:                        clock,
		killPodFunc:                  killPodFunc,
		mirrorPodFunc:                mirrorPodFunc,
		imageGC:                      imageGC,
		containerGC:                  containerGC,
		config:                       config,
		recorder:                     recorder,
		summaryProvider:              summaryProvider,
		nodeRef:                      nodeRef,
		nodeConditionsLastObservedAt: nodeConditionsObservedAt{},
		thresholdsFirstObservedAt:    thresholdsObservedAt{},
		dedicatedImageFs:             nil,
		thresholdNotifiers:           []ThresholdNotifier{},
		etcHostsPath:                 etcHostsPath,
	}

	return manager, manager
}
