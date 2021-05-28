package v1alpha1

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	schedutil "k8s-lx1036/k8s/scheduler/pkg/scheduler/util"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/sets"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
	v1helper "k8s.io/kubernetes/pkg/apis/core/v1/helper"
	"k8s.io/kubernetes/pkg/features"
)

// AffinityTerm is a processed version of v1.PodAffinityTerm.
type AffinityTerm struct {
	Namespaces  sets.String
	Selector    labels.Selector
	TopologyKey string
}

// WeightedAffinityTerm is a "processed" representation of v1.WeightedAffinityTerm.
type WeightedAffinityTerm struct {
	AffinityTerm
	Weight int32
}

func newAffinityTerm(pod *v1.Pod, term *v1.PodAffinityTerm) (*AffinityTerm, error) {
	namespaces := schedutil.GetNamespacesFromPodAffinityTerm(pod, term)
	selector, err := metav1.LabelSelectorAsSelector(term.LabelSelector)
	if err != nil {
		return nil, err
	}
	return &AffinityTerm{Namespaces: namespaces, Selector: selector, TopologyKey: term.TopologyKey}, nil
}

// getAffinityTerms receives a Pod and affinity terms and returns the namespaces and
// selectors of the terms.
func getAffinityTerms(pod *v1.Pod, v1Terms []v1.PodAffinityTerm) ([]AffinityTerm, error) {
	if v1Terms == nil {
		return nil, nil
	}

	var terms []AffinityTerm
	for _, term := range v1Terms {
		t, err := newAffinityTerm(pod, &term)
		if err != nil {
			// We get here if the label selector failed to process
			return nil, err
		}
		terms = append(terms, *t)
	}
	return terms, nil
}

// getWeightedAffinityTerms returns the list of processed affinity terms.
func getWeightedAffinityTerms(pod *v1.Pod, v1Terms []v1.WeightedPodAffinityTerm) ([]WeightedAffinityTerm, error) {
	if v1Terms == nil {
		return nil, nil
	}

	var terms []WeightedAffinityTerm
	for _, term := range v1Terms {
		t, err := newAffinityTerm(pod, &term.PodAffinityTerm)
		if err != nil {
			// We get here if the label selector failed to process
			return nil, err
		}
		terms = append(terms, WeightedAffinityTerm{AffinityTerm: *t, Weight: term.Weight})
	}
	return terms, nil
}

// PodInfo is a wrapper to a Pod with additional pre-computed information to
// accelerate processing. This information is typically immutable (e.g., pre-processed
// inter-pod affinity selectors).
type PodInfo struct {
	Pod                        *v1.Pod
	RequiredAffinityTerms      []AffinityTerm
	RequiredAntiAffinityTerms  []AffinityTerm
	PreferredAffinityTerms     []WeightedAffinityTerm
	PreferredAntiAffinityTerms []WeightedAffinityTerm
	ParseError                 error
}

// INFO: 抽取 pod affinity 相关信息
func NewPodInfo(pod *v1.Pod) *PodInfo {
	var preferredAffinityTerms []v1.WeightedPodAffinityTerm
	var preferredAntiAffinityTerms []v1.WeightedPodAffinityTerm
	if affinity := pod.Spec.Affinity; affinity != nil {
		if a := affinity.PodAffinity; a != nil {
			preferredAffinityTerms = a.PreferredDuringSchedulingIgnoredDuringExecution
		}
		if a := affinity.PodAntiAffinity; a != nil {
			preferredAntiAffinityTerms = a.PreferredDuringSchedulingIgnoredDuringExecution
		}
	}

	// Attempt to parse the affinity terms
	var parseErr error
	requiredAffinityTerms, err := getAffinityTerms(pod, schedutil.GetPodAffinityTerms(pod.Spec.Affinity))
	if err != nil {
		parseErr = fmt.Errorf("requiredAffinityTerms: %w", err)
	}
	requiredAntiAffinityTerms, err := getAffinityTerms(pod, schedutil.GetPodAntiAffinityTerms(pod.Spec.Affinity))
	if err != nil {
		parseErr = fmt.Errorf("requiredAntiAffinityTerms: %w", err)
	}
	weightedAffinityTerms, err := getWeightedAffinityTerms(pod, preferredAffinityTerms)
	if err != nil {
		parseErr = fmt.Errorf("preferredAffinityTerms: %w", err)
	}
	weightedAntiAffinityTerms, err := getWeightedAffinityTerms(pod, preferredAntiAffinityTerms)
	if err != nil {
		parseErr = fmt.Errorf("preferredAntiAffinityTerms: %w", err)
	}

	return &PodInfo{
		Pod:                        pod,
		RequiredAffinityTerms:      requiredAffinityTerms,
		RequiredAntiAffinityTerms:  requiredAntiAffinityTerms,
		PreferredAffinityTerms:     weightedAffinityTerms,
		PreferredAntiAffinityTerms: weightedAntiAffinityTerms,
		ParseError:                 parseErr,
	}
}

// NodeInfo is node level aggregated information.
type NodeInfo struct {
	// Overall node information.
	node *v1.Node

	// Pods running on the node.
	Pods []*PodInfo

	// The subset of pods with affinity.
	PodsWithAffinity []*PodInfo

	// The subset of pods with required anti-affinity.
	PodsWithRequiredAntiAffinity []*PodInfo

	// Ports allocated on the node.
	UsedPorts HostPortInfo

	// Total requested resources of all pods on this node. This includes assumed
	// pods, which scheduler has sent for binding, but may not be scheduled yet.
	Requested *Resource
	// Total requested resources of all pods on this node with a minimum value
	// applied to each container's CPU and memory requests. This does not reflect
	// the actual resource requests for this node, but is used to avoid scheduling
	// many zero-request pods onto one node.
	NonZeroRequested *Resource
	// We store allocatedResources (which is Node.Status.Allocatable.*) explicitly
	// as int64, to avoid conversions and accessing map.
	Allocatable *Resource

	// ImageStates holds the entry of an image if and only if this image is on the node. The entry can be used for
	// checking an image's existence and advanced usage (e.g., image locality scheduling policy) based on the image
	// state information.
	ImageStates map[string]*ImageStateSummary

	// TransientInfo holds the information pertaining to a scheduling cycle. This will be destructed at the end of
	// scheduling cycle.
	// INFO: @ravig. Remove this once we have a clear approach for message passing across predicates and priorities.
	TransientInfo *TransientSchedulerInfo

	// Whenever NodeInfo changes, generation is bumped.
	// This is used to avoid cloning it if the object didn't change.
	Generation int64
}

// INFO: 把 pod 加入当前 node，更新下 nodeInfo 相关信息
func (n *NodeInfo) AddPod(pod *v1.Pod) {
	podInfo := NewPodInfo(pod)
	res, non0CPU, non0Mem := calculateResource(pod)
	n.Requested.MilliCPU += res.MilliCPU
	n.Requested.Memory += res.Memory
	n.Requested.EphemeralStorage += res.EphemeralStorage
	if n.Requested.ScalarResources == nil && len(res.ScalarResources) > 0 {
		n.Requested.ScalarResources = map[v1.ResourceName]int64{}
	}
	for rName, rQuant := range res.ScalarResources {
		n.Requested.ScalarResources[rName] += rQuant
	}
	n.NonZeroRequested.MilliCPU += non0CPU
	n.NonZeroRequested.Memory += non0Mem
	if podWithAffinity(pod) {
		n.PodsWithAffinity = append(n.PodsWithAffinity, podInfo)
	}
	if podWithRequiredAntiAffinity(pod) {
		n.PodsWithRequiredAntiAffinity = append(n.PodsWithRequiredAntiAffinity, podInfo)
	}

	n.Pods = append(n.Pods, podInfo)

	// Consume ports when pods added.
	n.updateUsedPorts(podInfo.Pod, true)

	n.Generation = nextGeneration()
}

func podWithAffinity(p *v1.Pod) bool {
	affinity := p.Spec.Affinity
	return affinity != nil && (affinity.PodAffinity != nil || affinity.PodAntiAffinity != nil)
}
func podWithRequiredAntiAffinity(p *v1.Pod) bool {
	affinity := p.Spec.Affinity
	return affinity != nil && affinity.PodAntiAffinity != nil &&
		len(affinity.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution) != 0
}

// resourceRequest = max(sum(podSpec.Containers), podSpec.InitContainers) + overHead
func calculateResource(pod *v1.Pod) (res Resource, non0CPU int64, non0Mem int64) {
	resPtr := &res
	for _, c := range pod.Spec.Containers {
		resPtr.Add(c.Resources.Requests)
		non0CPUReq, non0MemReq := schedutil.GetNonzeroRequests(&c.Resources.Requests)
		non0CPU += non0CPUReq
		non0Mem += non0MemReq
		// No non-zero resources for GPUs or opaque resources.
	}

	for _, ic := range pod.Spec.InitContainers {
		resPtr.SetMaxResource(ic.Resources.Requests)
		non0CPUReq, non0MemReq := schedutil.GetNonzeroRequests(&ic.Resources.Requests)
		if non0CPU < non0CPUReq {
			non0CPU = non0CPUReq
		}

		if non0Mem < non0MemReq {
			non0Mem = non0MemReq
		}
	}

	// If Overhead is being utilized, add to the total requests for the pod
	if pod.Spec.Overhead != nil && utilfeature.DefaultFeatureGate.Enabled(features.PodOverhead) {
		resPtr.Add(pod.Spec.Overhead)
		if _, found := pod.Spec.Overhead[v1.ResourceCPU]; found {
			non0CPU += pod.Spec.Overhead.Cpu().MilliValue()
		}

		if _, found := pod.Spec.Overhead[v1.ResourceMemory]; found {
			non0Mem += pod.Spec.Overhead.Memory().Value()
		}
	}

	return
}

// updateUsedPorts updates the UsedPorts of NodeInfo.
func (n *NodeInfo) updateUsedPorts(pod *v1.Pod, add bool) {
	for j := range pod.Spec.Containers {
		container := &pod.Spec.Containers[j]
		for k := range container.Ports {
			podPort := &container.Ports[k]
			if add {
				n.UsedPorts.Add(podPort.HostIP, string(podPort.Protocol), podPort.HostPort)
			} else {
				n.UsedPorts.Remove(podPort.HostIP, string(podPort.Protocol), podPort.HostPort)
			}
		}
	}
}

// RemovePod subtracts pod information from this NodeInfo.
func (n *NodeInfo) RemovePod(pod *v1.Pod) error {
	// TODO:
	return nil
}

// SetNode sets the overall node information.
func (n *NodeInfo) SetNode(node *v1.Node) error {
	n.node = node
	n.Allocatable = NewResource(node.Status.Allocatable)
	n.TransientInfo = NewTransientSchedulerInfo()
	n.Generation = nextGeneration()
	return nil
}

// Node returns overall information about this node.
func (n *NodeInfo) Node() *v1.Node {
	if n == nil {
		return nil
	}
	return n.node
}

// NewNodeInfo returns a ready to use empty NodeInfo object.
// If any pods are given in arguments, their information will be aggregated in
// the returned object.
func NewNodeInfo(pods ...*v1.Pod) *NodeInfo {
	ni := &NodeInfo{
		Requested:        &Resource{},
		NonZeroRequested: &Resource{},
		Allocatable:      &Resource{},
		TransientInfo:    NewTransientSchedulerInfo(),
		Generation:       nextGeneration(),
		UsedPorts:        make(HostPortInfo),
		ImageStates:      make(map[string]*ImageStateSummary),
	}
	for _, pod := range pods {
		ni.AddPod(pod)
	}
	return ni
}

// Resource is a collection of compute resource.
type Resource struct {
	MilliCPU         int64
	Memory           int64
	EphemeralStorage int64
	// We store allowedPodNumber (which is Node.Status.Allocatable.Pods().Value())
	// explicitly as int, to avoid conversions and improve performance.
	AllowedPodNumber int
	// ScalarResources
	ScalarResources map[v1.ResourceName]int64
}

func (r *Resource) ResourceList() v1.ResourceList {
	result := v1.ResourceList{
		v1.ResourceCPU:              *resource.NewMilliQuantity(r.MilliCPU, resource.DecimalSI),
		v1.ResourceMemory:           *resource.NewQuantity(r.Memory, resource.BinarySI),
		v1.ResourcePods:             *resource.NewQuantity(int64(r.AllowedPodNumber), resource.BinarySI),
		v1.ResourceEphemeralStorage: *resource.NewQuantity(r.EphemeralStorage, resource.BinarySI),
	}
	for rName, rQuant := range r.ScalarResources {
		if v1helper.IsHugePageResourceName(rName) {
			result[rName] = *resource.NewQuantity(rQuant, resource.BinarySI)
		} else {
			result[rName] = *resource.NewQuantity(rQuant, resource.DecimalSI)
		}
	}
	return result
}

// SetMaxResource compares with ResourceList and takes max value for each Resource.
func (r *Resource) SetMaxResource(rl v1.ResourceList) {
	if r == nil {
		return
	}

	for rName, rQuantity := range rl {
		switch rName {
		case v1.ResourceMemory:
			if mem := rQuantity.Value(); mem > r.Memory {
				r.Memory = mem
			}
		case v1.ResourceCPU:
			if cpu := rQuantity.MilliValue(); cpu > r.MilliCPU {
				r.MilliCPU = cpu
			}
		case v1.ResourceEphemeralStorage:
			if utilfeature.DefaultFeatureGate.Enabled(features.LocalStorageCapacityIsolation) {
				if ephemeralStorage := rQuantity.Value(); ephemeralStorage > r.EphemeralStorage {
					r.EphemeralStorage = ephemeralStorage
				}
			}
		default:
			if v1helper.IsScalarResourceName(rName) {
				value := rQuantity.Value()
				if value > r.ScalarResources[rName] {
					r.SetScalar(rName, value)
				}
			}
		}
	}
}

func (r *Resource) Add(rl v1.ResourceList) {
	if r == nil {
		return
	}

	for rName, rQuant := range rl {
		switch rName {
		case v1.ResourceCPU:
			r.MilliCPU += rQuant.MilliValue()
		case v1.ResourceMemory:
			r.Memory += rQuant.Value()
		case v1.ResourcePods:
			r.AllowedPodNumber += int(rQuant.Value())
		case v1.ResourceEphemeralStorage:
			if utilfeature.DefaultFeatureGate.Enabled(features.LocalStorageCapacityIsolation) {
				// if the local storage capacity isolation feature gate is disabled, pods request 0 disk.
				r.EphemeralStorage += rQuant.Value()
			}
		default:
			if v1helper.IsScalarResourceName(rName) {
				r.AddScalar(rName, rQuant.Value())
			}
		}
	}
}

// AddScalar adds a resource by a scalar value of this resource.
func (r *Resource) AddScalar(name v1.ResourceName, quantity int64) {
	r.SetScalar(name, r.ScalarResources[name]+quantity)
}

// SetScalar sets a resource by a scalar value of this resource.
func (r *Resource) SetScalar(name v1.ResourceName, quantity int64) {
	// Lazily allocate scalar resource map.
	if r.ScalarResources == nil {
		r.ScalarResources = map[v1.ResourceName]int64{}
	}
	r.ScalarResources[name] = quantity
}

// NewResource creates a Resource from ResourceList
func NewResource(rl v1.ResourceList) *Resource {
	r := &Resource{}
	r.Add(rl)
	return r
}

// ImageStateSummary provides summarized information about the state of an image.
type ImageStateSummary struct {
	// Size of the image
	Size int64
	// Used to track how many nodes have this image
	NumNodes int
}

// nodeTransientInfo contains transient node information while scheduling.
type nodeTransientInfo struct {
	// AllocatableVolumesCount contains number of volumes that could be attached to node.
	AllocatableVolumesCount int
	// Requested number of volumes on a particular node.
	RequestedVolumes int
}

//initializeNodeTransientInfo initializes transient information pertaining to node.
func initializeNodeTransientInfo() nodeTransientInfo {
	return nodeTransientInfo{AllocatableVolumesCount: 0, RequestedVolumes: 0}
}

// TransientSchedulerInfo is a transient structure which is destructed at the end of each scheduling cycle.
// It consists of items that are valid for a scheduling cycle and is used for message passing across predicates and
// priorities. Some examples which could be used as fields are number of volumes being used on node, current utilization
// on node etc.
// IMPORTANT NOTE: Make sure that each field in this structure is documented along with usage. Expand this structure
// only when absolutely needed as this data structure will be created and destroyed during every scheduling cycle.
type TransientSchedulerInfo struct {
	TransientLock sync.Mutex
	// NodeTransInfo holds the information related to nodeTransientInformation. NodeName is the key here.
	TransNodeInfo nodeTransientInfo
}

// NewTransientSchedulerInfo returns a new scheduler transient structure with initialized values.
func NewTransientSchedulerInfo() *TransientSchedulerInfo {
	tsi := &TransientSchedulerInfo{
		TransNodeInfo: initializeNodeTransientInfo(),
	}
	return tsi
}

var generation int64

// nextGeneration: Let's make sure history never forgets the name...
// Increments the generation number monotonically ensuring that generation numbers never collide.
// Collision of the generation numbers would be particularly problematic if a node was deleted and
// added back with the same name. See issue#63262.
func nextGeneration() int64 {
	return atomic.AddInt64(&generation, 1)
}

// QueuedPodInfo is a Pod wrapper with additional information related to
// the pod's status in the scheduling queue, such as the timestamp when
// it's added to the queue.
type QueuedPodInfo struct {
	Pod *v1.Pod
	// The time pod added to the scheduling queue.
	Timestamp time.Time
	// Number of schedule attempts before successfully scheduled.
	// It's used to record the # attempts metric.
	Attempts int
	// The time when the pod is added to the queue for the first time. The pod may be added
	// back to the queue multiple times before it's successfully scheduled.
	// It shouldn't be updated once initialized. It's used to record the e2e scheduling
	// latency for a pod.
	InitialAttemptTimestamp time.Time
}

// DeepCopy returns a deep copy of the QueuedPodInfo object.
func (pqi *QueuedPodInfo) DeepCopy() *QueuedPodInfo {
	return &QueuedPodInfo{
		Pod:                     pqi.Pod.DeepCopy(),
		Timestamp:               pqi.Timestamp,
		Attempts:                pqi.Attempts,
		InitialAttemptTimestamp: pqi.InitialAttemptTimestamp,
	}
}

// DefaultBindAllHostIP defines the default ip address used to bind to all host.
const DefaultBindAllHostIP = "0.0.0.0"

// ProtocolPort represents a protocol port pair, e.g. tcp:80.
type ProtocolPort struct {
	Protocol string
	Port     int32
}

// NewProtocolPort creates a ProtocolPort instance.
func NewProtocolPort(protocol string, port int32) *ProtocolPort {
	pp := &ProtocolPort{
		Protocol: protocol,
		Port:     port,
	}

	if len(pp.Protocol) == 0 {
		pp.Protocol = string(v1.ProtocolTCP)
	}

	return pp
}

// HostPortInfo stores mapping from ip to a set of ProtocolPort
type HostPortInfo map[string]map[ProtocolPort]struct{}

// CheckConflict checks if the input (ip, protocol, port) conflicts with the existing
// ones in HostPortInfo.
func (h HostPortInfo) CheckConflict(ip, protocol string, port int32) bool {
	if port <= 0 {
		return false
	}

	h.sanitize(&ip, &protocol)

	pp := NewProtocolPort(protocol, port)

	// If ip is 0.0.0.0 check all IP's (protocol, port) pair
	if ip == DefaultBindAllHostIP {
		for _, m := range h {
			if _, ok := m[*pp]; ok {
				return true
			}
		}
		return false
	}

	// If ip isn't 0.0.0.0, only check IP and 0.0.0.0's (protocol, port) pair
	for _, key := range []string{DefaultBindAllHostIP, ip} {
		if m, ok := h[key]; ok {
			if _, ok2 := m[*pp]; ok2 {
				return true
			}
		}
	}

	return false
}

// sanitize the parameters
func (h HostPortInfo) sanitize(ip, protocol *string) {
	if len(*ip) == 0 {
		*ip = DefaultBindAllHostIP
	}
	if len(*protocol) == 0 {
		*protocol = string(v1.ProtocolTCP)
	}
}

// Add adds (ip, protocol, port) to HostPortInfo
func (h HostPortInfo) Add(ip, protocol string, port int32) {
	if port <= 0 {
		return
	}

	h.sanitize(&ip, &protocol)

	pp := NewProtocolPort(protocol, port)
	if _, ok := h[ip]; !ok {
		h[ip] = map[ProtocolPort]struct{}{
			*pp: {},
		}
		return
	}

	h[ip][*pp] = struct{}{}
}

// Remove removes (ip, protocol, port) from HostPortInfo
func (h HostPortInfo) Remove(ip, protocol string, port int32) {
	if port <= 0 {
		return
	}

	h.sanitize(&ip, &protocol)

	pp := NewProtocolPort(protocol, port)
	if m, ok := h[ip]; ok {
		delete(m, *pp)
		if len(h[ip]) == 0 {
			delete(h, ip)
		}
	}
}
