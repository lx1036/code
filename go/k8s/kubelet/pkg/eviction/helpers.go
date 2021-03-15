package eviction

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	statsapi "k8s-lx1036/k8s/kubelet/pkg/apis/v1alpha1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1helpers "k8s.io/component-helpers/scheduling/corev1"
	"k8s.io/klog/v2"
	v1resource "k8s.io/kubernetes/pkg/api/v1/resource"
)

const (
	unsupportedEvictionSignal = "unsupported eviction signal %v"
	// Reason is the reason reported back in status.
	Reason = "Evicted"
	// nodeLowMessageFmt is the message for evictions due to resource pressure.
	nodeLowMessageFmt = "The node was low on resource: %v. "
	// nodeConditionMessageFmt is the message for evictions due to resource pressure.
	nodeConditionMessageFmt = "The node had condition: %v. "
	// containerMessageFmt provides additional information for containers exceeding requests
	containerMessageFmt = "Container %s was using %s, which exceeds its request of %s. "
	// containerEphemeralStorageMessageFmt provides additional information for containers which have exceeded their ES limit
	containerEphemeralStorageMessageFmt = "Container %s exceeded its local ephemeral storage limit %q. "
	// podEphemeralStorageMessageFmt provides additional information for pods which have exceeded their ES limit
	podEphemeralStorageMessageFmt = "Pod ephemeral local storage usage exceeds the total limit of containers %s. "
	// emptyDirMessageFmt provides additional information for empty-dir volumes which have exceeded their size limit
	emptyDirMessageFmt = "Usage of EmptyDir volume %q exceeds the limit %q. "
	// inodes, number. internal to this module, used to account for local disk inode consumption.
	resourceInodes v1.ResourceName = "inodes"
	// resourcePids, number. internal to this module, used to account for local pid consumption.
	resourcePids v1.ResourceName = "pids"
	// OffendingContainersKey is the key in eviction event annotations for the list of container names which exceeded their requests
	OffendingContainersKey = "offending_containers"
	// OffendingContainersUsageKey is the key in eviction event annotations for the list of usage of containers which exceeded their requests
	OffendingContainersUsageKey = "offending_containers_usage"
	// StarvedResourceKey is the key for the starved resource in eviction event annotations
	StarvedResourceKey = "starved_resource"
)

const (
	// User visible keys for managing node allocatable enforcement on the node.
	NodeAllocatableEnforcementKey = "pods"
	SystemReservedEnforcementKey  = "system-reserved"
	KubeReservedEnforcementKey    = "kube-reserved"
	NodeAllocatableNoneKey        = "none"
)

var (
	// signalToNodeCondition maps a signal to the node condition to report if threshold is met.
	signalToNodeCondition map[Signal]v1.NodeConditionType
	// signalToResource maps a Signal to its associated Resource.
	signalToResource map[Signal]v1.ResourceName
)

func init() {
	// map eviction signals to node conditions
	signalToNodeCondition = map[Signal]v1.NodeConditionType{}
	signalToNodeCondition[SignalMemoryAvailable] = v1.NodeMemoryPressure
	signalToNodeCondition[SignalAllocatableMemoryAvailable] = v1.NodeMemoryPressure
	signalToNodeCondition[SignalImageFsAvailable] = v1.NodeDiskPressure
	signalToNodeCondition[SignalNodeFsAvailable] = v1.NodeDiskPressure
	signalToNodeCondition[SignalImageFsInodesFree] = v1.NodeDiskPressure
	signalToNodeCondition[SignalNodeFsInodesFree] = v1.NodeDiskPressure
	signalToNodeCondition[SignalPIDAvailable] = v1.NodePIDPressure

	// map signals to resources (and vice-versa)
	signalToResource = map[Signal]v1.ResourceName{}
	signalToResource[SignalMemoryAvailable] = v1.ResourceMemory
	signalToResource[SignalAllocatableMemoryAvailable] = v1.ResourceMemory
	signalToResource[SignalImageFsAvailable] = v1.ResourceEphemeralStorage
	signalToResource[SignalImageFsInodesFree] = resourceInodes
	signalToResource[SignalNodeFsAvailable] = v1.ResourceEphemeralStorage
	signalToResource[SignalNodeFsInodesFree] = resourceInodes
	signalToResource[SignalPIDAvailable] = resourcePids
}

// parseGracePeriods parses the grace period statements
func parseGracePeriods(statements map[string]string) (map[Signal]time.Duration, error) {
	if len(statements) == 0 {
		return nil, nil
	}
	results := map[Signal]time.Duration{}
	for signal, val := range statements {
		signal := Signal(signal)
		if !validSignal(signal) {
			return nil, fmt.Errorf(unsupportedEvictionSignal, signal)
		}
		gracePeriod, err := time.ParseDuration(val)
		if err != nil {
			return nil, err
		}
		if gracePeriod < 0 {
			return nil, fmt.Errorf("invalid eviction grace period specified: %v, must be a positive value", val)
		}
		results[signal] = gracePeriod
	}
	return results, nil
}

// parseMinimumReclaims parses the minimum reclaim statements
func parseMinimumReclaims(statements map[string]string) (map[Signal]ThresholdValue, error) {
	if len(statements) == 0 {
		return nil, nil
	}
	results := map[Signal]ThresholdValue{}
	for signal, val := range statements {
		signal := Signal(signal)
		if !validSignal(signal) {
			return nil, fmt.Errorf(unsupportedEvictionSignal, signal)
		}
		if strings.HasSuffix(val, "%") {
			percentage, err := parsePercentage(val)
			if err != nil {
				return nil, err
			}
			if percentage <= 0 {
				return nil, fmt.Errorf("eviction percentage minimum reclaim %v must be positive: %s", signal, val)
			}
			results[signal] = ThresholdValue{
				Percentage: percentage,
			}
			continue
		}
		quantity, err := resource.ParseQuantity(val)
		if err != nil {
			return nil, err
		}
		if quantity.Sign() < 0 {
			return nil, fmt.Errorf("negative eviction minimum reclaim specified for %v", signal)
		}
		results[signal] = ThresholdValue{
			Quantity: &quantity,
		}
	}
	return results, nil
}

// isHardEvictionThreshold returns true if eviction should immediately occur
func isHardEvictionThreshold(threshold Threshold) bool {
	return threshold.GracePeriod == time.Duration(0)
}
func addAllocatableThresholds(thresholds []Threshold) []Threshold {
	additionalThresholds := []Threshold{}
	for _, threshold := range thresholds {
		if threshold.Signal == SignalMemoryAvailable && isHardEvictionThreshold(threshold) {
			// Copy the SignalMemoryAvailable to SignalAllocatableMemoryAvailable
			additionalThresholds = append(additionalThresholds, Threshold{
				Signal:     SignalAllocatableMemoryAvailable,
				Operator:   threshold.Operator,
				Value:      threshold.Value,
				MinReclaim: threshold.MinReclaim,
			})
		}
	}
	return append(append([]Threshold{}, thresholds...), additionalThresholds...)
}

// ParseThresholdConfig parses the flags for thresholds.
func ParseThresholdConfig(allocatableConfig []string, evictionHard, evictionSoft, evictionSoftGracePeriod, evictionMinimumReclaim map[string]string) ([]Threshold, error) {
	var results []Threshold
	hardThresholds, err := parseThresholdStatements(evictionHard)
	if err != nil {
		return nil, err
	}
	results = append(results, hardThresholds...)
	softThresholds, err := parseThresholdStatements(evictionSoft)
	if err != nil {
		return nil, err
	}
	gracePeriods, err := parseGracePeriods(evictionSoftGracePeriod)
	if err != nil {
		return nil, err
	}
	minReclaims, err := parseMinimumReclaims(evictionMinimumReclaim)
	if err != nil {
		return nil, err
	}
	for i := range softThresholds {
		signal := softThresholds[i].Signal
		period, found := gracePeriods[signal]
		if !found {
			return nil, fmt.Errorf("grace period must be specified for the soft eviction threshold %v", signal)
		}
		softThresholds[i].GracePeriod = period
	}
	results = append(results, softThresholds...)
	for i := range results {
		if minReclaim, ok := minReclaims[results[i].Signal]; ok {
			results[i].MinReclaim = &minReclaim
		}
	}
	for _, key := range allocatableConfig {
		if key == NodeAllocatableEnforcementKey {
			results = addAllocatableThresholds(results)
			break
		}
	}
	return results, nil
}

// parseThresholdStatements parses the input statements into a list of Threshold objects.
func parseThresholdStatements(statements map[string]string) ([]Threshold, error) {
	if len(statements) == 0 {
		return nil, nil
	}
	var results []Threshold
	for signal, val := range statements {
		result, err := parseThresholdStatement(Signal(signal), val)
		if err != nil {
			return nil, err
		}
		if result != nil {
			results = append(results, *result)
		}
	}
	return results, nil
}

// validSignal returns true if the signal is supported.
func validSignal(signal Signal) bool {
	_, found := signalToResource[signal]
	return found
}

// parsePercentage parses a string representing a percentage value
func parsePercentage(input string) (float32, error) {
	value, err := strconv.ParseFloat(strings.TrimRight(input, "%"), 32)
	if err != nil {
		return 0, err
	}
	return float32(value) / 100, nil
}

// parseThresholdStatement parses a threshold statement and returns a threshold,
// or nil if the threshold should be ignored.
func parseThresholdStatement(signal Signal, val string) (*Threshold, error) {
	if !validSignal(signal) {
		return nil, fmt.Errorf(unsupportedEvictionSignal, signal)
	}
	operator := OpForSignal[signal]
	if strings.HasSuffix(val, "%") {
		// ignore 0% and 100%
		if val == "0%" || val == "100%" {
			return nil, nil
		}
		percentage, err := parsePercentage(val)
		if err != nil {
			return nil, err
		}
		if percentage < 0 {
			return nil, fmt.Errorf("eviction percentage threshold %v must be >= 0%%: %s", signal, val)
		}
		if percentage > 100 {
			return nil, fmt.Errorf("eviction percentage threshold %v must be <= 100%%: %s", signal, val)
		}
		return &Threshold{
			Signal:   signal,
			Operator: operator,
			Value: ThresholdValue{
				Percentage: percentage,
			},
		}, nil
	}
	quantity, err := resource.ParseQuantity(val)
	if err != nil {
		return nil, err
	}
	if quantity.Sign() < 0 || quantity.IsZero() {
		return nil, fmt.Errorf("eviction threshold %v must be positive: %s", signal, &quantity)
	}
	return &Threshold{
		Signal:   signal,
		Operator: operator,
		Value: ThresholdValue{
			Quantity: &quantity,
		},
	}, nil
}

// hasNodeCondition returns true if the node condition is in the input list
func hasNodeCondition(inputs []v1.NodeConditionType, item v1.NodeConditionType) bool {
	for _, input := range inputs {
		if input == item {
			return true
		}
	}
	return false
}

// signalObservation is the observed resource usage
type signalObservation struct {
	// The resource capacity
	capacity *resource.Quantity
	// The available resource
	available *resource.Quantity
	// Time at which the observation was taken
	time metav1.Time
}

// signalObservations maps a signal to an observed quantity
type signalObservations map[Signal]signalObservation

// makeSignalObservations derives observations using the specified summary provider.
func makeSignalObservations(summary *statsapi.Summary) (signalObservations, statsFunc) {
	// build an evaluation context for current eviction signals
	result := signalObservations{}
	if memory := summary.Node.Memory; memory != nil && memory.AvailableBytes != nil && memory.WorkingSetBytes != nil {
		result[SignalMemoryAvailable] = signalObservation{
			available: resource.NewQuantity(int64(*memory.AvailableBytes), resource.BinarySI),
			capacity:  resource.NewQuantity(int64(*memory.AvailableBytes+*memory.WorkingSetBytes), resource.BinarySI),
			time:      memory.Time,
		}
	}

	if allocatableContainer, err := getSysContainer(summary.Node.SystemContainers, statsapi.SystemContainerPods); err != nil {
		klog.Errorf("eviction manager: failed to construct signal: %q error: %v", SignalAllocatableMemoryAvailable, err)
	} else {
		if memory := allocatableContainer.Memory; memory != nil && memory.AvailableBytes != nil && memory.WorkingSetBytes != nil {
			result[SignalAllocatableMemoryAvailable] = signalObservation{
				available: resource.NewQuantity(int64(*memory.AvailableBytes), resource.BinarySI),
				capacity:  resource.NewQuantity(int64(*memory.AvailableBytes+*memory.WorkingSetBytes), resource.BinarySI),
				time:      memory.Time,
			}
		}
	}

	if nodeFs := summary.Node.Fs; nodeFs != nil {
		if nodeFs.AvailableBytes != nil && nodeFs.CapacityBytes != nil {
			result[SignalNodeFsAvailable] = signalObservation{
				available: resource.NewQuantity(int64(*nodeFs.AvailableBytes), resource.BinarySI),
				capacity:  resource.NewQuantity(int64(*nodeFs.CapacityBytes), resource.BinarySI),
				time:      nodeFs.Time,
			}
		}
		if nodeFs.InodesFree != nil && nodeFs.Inodes != nil {
			result[SignalNodeFsInodesFree] = signalObservation{
				available: resource.NewQuantity(int64(*nodeFs.InodesFree), resource.DecimalSI),
				capacity:  resource.NewQuantity(int64(*nodeFs.Inodes), resource.DecimalSI),
				time:      nodeFs.Time,
			}
		}
	}

	if summary.Node.Runtime != nil {
		if imageFs := summary.Node.Runtime.ImageFs; imageFs != nil {
			if imageFs.AvailableBytes != nil && imageFs.CapacityBytes != nil {
				result[SignalImageFsAvailable] = signalObservation{
					available: resource.NewQuantity(int64(*imageFs.AvailableBytes), resource.BinarySI),
					capacity:  resource.NewQuantity(int64(*imageFs.CapacityBytes), resource.BinarySI),
					time:      imageFs.Time,
				}
				if imageFs.InodesFree != nil && imageFs.Inodes != nil {
					result[SignalImageFsInodesFree] = signalObservation{
						available: resource.NewQuantity(int64(*imageFs.InodesFree), resource.DecimalSI),
						capacity:  resource.NewQuantity(int64(*imageFs.Inodes), resource.DecimalSI),
						time:      imageFs.Time,
					}
				}
			}
		}
	}

	if rlimit := summary.Node.Rlimit; rlimit != nil {
		if rlimit.NumOfRunningProcesses != nil && rlimit.MaxPID != nil {
			available := int64(*rlimit.MaxPID) - int64(*rlimit.NumOfRunningProcesses)
			result[SignalPIDAvailable] = signalObservation{
				available: resource.NewQuantity(available, resource.DecimalSI),
				capacity:  resource.NewQuantity(int64(*rlimit.MaxPID), resource.DecimalSI),
				time:      rlimit.Time,
			}
		}
	}

	// build the function to work against for pod stats
	statsFunc := cachedStatsFunc(summary.Pods)
	return result, statsFunc
}

// cachedStatsFunc returns a statsFunc based on the provided pod stats.
func cachedStatsFunc(podStats []statsapi.PodStats) statsFunc {
	uid2PodStats := map[string]statsapi.PodStats{}
	for i := range podStats {
		uid2PodStats[podStats[i].PodRef.UID] = podStats[i]
	}
	return func(pod *v1.Pod) (statsapi.PodStats, bool) {
		stats, found := uid2PodStats[string(pod.UID)]
		return stats, found
	}
}

// thresholdsMet returns the set of thresholds that were met independent of grace period
// 比较实际观察值observations，和限定条件thresholds，判断是否有threshold满足条件了
func thresholdsMet(thresholds []Threshold, observations signalObservations, enforceMinReclaim bool) []Threshold {
	var results []Threshold
	for i := range thresholds {
		threshold := thresholds[i]
		observed, found := observations[threshold.Signal]
		if !found {
			klog.Warningf("eviction manager: no observation found for eviction signal %v", threshold.Signal)
			continue
		}
		// determine if we have met the specified threshold
		thresholdMet := false
		quantity := GetThresholdQuantity(threshold.Value, observed.capacity)
		// if enforceMinReclaim is specified, we compare relative to value - minreclaim
		if enforceMinReclaim && threshold.MinReclaim != nil {
			quantity.Add(*GetThresholdQuantity(*threshold.MinReclaim, observed.capacity))
		}
		thresholdResult := quantity.Cmp(*observed.available)
		switch threshold.Operator {
		case OpLessThan:
			thresholdMet = thresholdResult > 0
		}
		if thresholdMet {
			results = append(results, threshold)
		}
	}

	return results
}

// getReclaimableThreshold finds the threshold and resource to reclaim
func getReclaimableThreshold(thresholds []Threshold) (Threshold, v1.ResourceName, bool) {
	for _, thresholdToReclaim := range thresholds {
		if resourceToReclaim, ok := signalToResource[thresholdToReclaim.Signal]; ok {
			return thresholdToReclaim, resourceToReclaim, true
		}
		klog.V(3).Infof("eviction manager: threshold %s was crossed, but reclaim is not implemented for this threshold.", thresholdToReclaim.Signal)
	}
	return Threshold{}, "", false
}

// byEvictionPriority implements sort.Interface for []v1.ResourceName.
type byEvictionPriority []Threshold

func (a byEvictionPriority) Len() int      { return len(a) }
func (a byEvictionPriority) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

// Less ranks memory before all other resources, and ranks thresholds with no resource to reclaim last
func (a byEvictionPriority) Less(i, j int) bool {
	_, jSignalHasResource := signalToResource[a[j].Signal]
	return a[i].Signal == SignalMemoryAvailable || a[i].Signal == SignalAllocatableMemoryAvailable || !jSignalHasResource
}

// Cmp compares p1 and p2 and returns:
//
//   -1 if p1 <  p2
//    0 if p1 == p2
//   +1 if p1 >  p2
//
type cmpFunc func(p1, p2 *v1.Pod) int

// multiSorter implements the Sort interface, sorting changes within.
type multiSorter struct {
	pods []*v1.Pod
	cmp  []cmpFunc
}

// Sort sorts the argument slice according to the less functions passed to OrderedBy.
func (ms *multiSorter) Sort(pods []*v1.Pod) {
	ms.pods = pods
	sort.Sort(ms)
}

// OrderedBy returns a Sorter that sorts using the cmp functions, in order.
// Call its Sort method to sort the data.
func orderedBy(cmp ...cmpFunc) *multiSorter {
	return &multiSorter{
		cmp: cmp,
	}
}

// Len is part of sort.Interface.
func (ms *multiSorter) Len() int {
	return len(ms.pods)
}

// Swap is part of sort.Interface.
func (ms *multiSorter) Swap(i, j int) {
	ms.pods[i], ms.pods[j] = ms.pods[j], ms.pods[i]
}

// Less is part of sort.Interface.
func (ms *multiSorter) Less(i, j int) bool {
	p1, p2 := ms.pods[i], ms.pods[j]
	var k int
	for k = 0; k < len(ms.cmp)-1; k++ {
		cmpResult := ms.cmp[k](p1, p2)
		// p1 is less than p2
		if cmpResult < 0 {
			return true
		}
		// p1 is greater than p2
		if cmpResult > 0 {
			return false
		}
		// we don't know yet
	}
	// the last cmp func is the final decider
	return ms.cmp[k](p1, p2) < 0
}

// priority compares pods by Priority, if priority is enabled.
func priority(p1, p2 *v1.Pod) int {
	priority1 := corev1helpers.PodPriority(p1)
	priority2 := corev1helpers.PodPriority(p2)
	if priority1 == priority2 {
		return 0
	}
	if priority1 > priority2 {
		return 1
	}
	return -1
}

// exceedMemoryRequests compares whether or not pods' memory usage exceeds their requests
func exceedMemoryRequests(stats statsFunc) cmpFunc {
	return func(p1, p2 *v1.Pod) int {
		p1Stats, p1Found := stats(p1)
		p2Stats, p2Found := stats(p2)
		if !p1Found || !p2Found {
			// prioritize evicting the pod for which no stats were found
			return cmpBool(!p1Found, !p2Found)
		}

		p1Memory := memoryUsage(p1Stats.Memory)
		p2Memory := memoryUsage(p2Stats.Memory)
		p1ExceedsRequests := p1Memory.Cmp(v1resource.GetResourceRequestQuantity(p1, v1.ResourceMemory)) == 1
		p2ExceedsRequests := p2Memory.Cmp(v1resource.GetResourceRequestQuantity(p2, v1.ResourceMemory)) == 1
		// prioritize evicting the pod which exceeds its requests
		return cmpBool(p1ExceedsRequests, p2ExceedsRequests)
	}
}

// memory compares pods by largest consumer of memory relative to request.
func memory(stats statsFunc) cmpFunc {
	return func(p1, p2 *v1.Pod) int {
		p1Stats, p1Found := stats(p1)
		p2Stats, p2Found := stats(p2)
		if !p1Found || !p2Found {
			// prioritize evicting the pod for which no stats were found
			return cmpBool(!p1Found, !p2Found)
		}

		// adjust p1, p2 usage relative to the request (if any)
		p1Memory := memoryUsage(p1Stats.Memory)
		p1Request := v1resource.GetResourceRequestQuantity(p1, v1.ResourceMemory)
		p1Memory.Sub(p1Request)

		p2Memory := memoryUsage(p2Stats.Memory)
		p2Request := v1resource.GetResourceRequestQuantity(p2, v1.ResourceMemory)
		p2Memory.Sub(p2Request)

		// prioritize evicting the pod which has the larger consumption of memory
		return p2Memory.Cmp(*p1Memory)
	}
}

// exceedDiskRequests compares whether or not pods' disk usage exceeds their requests
func exceedDiskRequests(stats statsFunc, fsStatsToMeasure []fsStatsType, diskResource v1.ResourceName) cmpFunc {
	return func(p1, p2 *v1.Pod) int {
		p1Stats, p1Found := stats(p1)
		p2Stats, p2Found := stats(p2)
		if !p1Found || !p2Found {
			// prioritize evicting the pod for which no stats were found
			return cmpBool(!p1Found, !p2Found)
		}

		p1Usage, p1Err := podDiskUsage(p1Stats, p1, fsStatsToMeasure)
		p2Usage, p2Err := podDiskUsage(p2Stats, p2, fsStatsToMeasure)
		if p1Err != nil || p2Err != nil {
			// prioritize evicting the pod which had an error getting stats
			return cmpBool(p1Err != nil, p2Err != nil)
		}

		p1Disk := p1Usage[diskResource]
		p2Disk := p2Usage[diskResource]
		p1ExceedsRequests := p1Disk.Cmp(v1resource.GetResourceRequestQuantity(p1, diskResource)) == 1
		p2ExceedsRequests := p2Disk.Cmp(v1resource.GetResourceRequestQuantity(p2, diskResource)) == 1
		// prioritize evicting the pod which exceeds its requests
		return cmpBool(p1ExceedsRequests, p2ExceedsRequests)
	}
}
