package api

import (
	"fmt"
	"time"

	v1 "k8s.io/api/core/v1"
	k8sframework "k8s.io/kubernetes/pkg/scheduler/framework"
)

type NodeInfo struct {
	Name string
	Node *v1.Node

	Tasks map[TaskID]*TaskInfo

	Allocatable   *Resource
	Capacity      *Resource
	ResourceUsage *NodeUsage
	// The idle resource on that node
	Idle *Resource
	// The used resource on that node, including running and terminating
	// pods
	Used *Resource
	// The releasing resource on that node
	Releasing *Resource
	// The pipelined resource on that node
	Pipelined *Resource

	// enable node resource oversubscription
	OversubscriptionNode bool
	// OfflineJobEvicting true means node resource usage too high then dispatched pod can not use oversubscription resource
	OfflineJobEvicting bool

	// The state of node
	State NodeState

	// Resource Oversubscription feature: the Oversubscription Resource reported in annotation
	OversubscriptionResource *Resource
	// ImageStates holds the entry of an image if and only if this image is on the node. The entry can be used for
	// checking an image's existence and advanced usage (e.g., image locality scheduling policy) based on the image
	// state information.
	ImageStates map[string]*k8sframework.ImageStateSummary

	// Used to store custom information
	Others map[string]interface{}
}

func (ni NodeInfo) String() string {
	tasks := ""

	i := 0
	for _, task := range ni.Tasks {
		tasks += fmt.Sprintf("\n\t %d: %v", i, task)
		i++
	}

	return fmt.Sprintf("Node (%s): allocatable<%v> idle <%v>, used <%v>, releasing <%v>, oversubscribution <%v>, "+
		"state <phase %s, reaseon %s>, oversubscributionNode <%v>, offlineJobEvicting <%v>,taints <%v>%s, imageStates %v",
		ni.Name, ni.Allocatable, ni.Idle, ni.Used, ni.Releasing, ni.OversubscriptionResource, ni.State.Phase, ni.State.Reason,
		ni.OversubscriptionNode, ni.OfflineJobEvicting, ni.Node.Spec.Taints, tasks, ni.ImageStates)
}

type NodeState struct {
	Phase  NodePhase
	Reason string
}

// NodeUsage defines the real load usage of node
type NodeUsage struct {
	MetricsTime time.Time
	CPUUsageAvg map[string]float64
	MEMUsageAvg map[string]float64
}

type CSINodeStatusInfo struct {
	CSINodeName  string
	DriverStatus map[string]bool
}

func NewNodeInfo(node *v1.Node) *NodeInfo {
	nodeInfo := &NodeInfo{
		Releasing: EmptyResource(),
		Pipelined: EmptyResource(),
		Idle:      EmptyResource(),
		Used:      EmptyResource(),

		Allocatable:   EmptyResource(),
		Capacity:      EmptyResource(),
		ResourceUsage: &NodeUsage{},

		OversubscriptionResource: EmptyResource(),
		Tasks:                    make(map[TaskID]*TaskInfo),

		Others:      make(map[string]interface{}),
		ImageStates: make(map[string]*k8sframework.ImageStateSummary),
	}

	nodeInfo.setOversubscription(node)

	if node != nil {
		nodeInfo.Name = node.Name
		nodeInfo.Node = node
		nodeInfo.Idle = NewResource(node.Status.Allocatable).Add(nodeInfo.OversubscriptionResource)
		nodeInfo.Allocatable = NewResource(node.Status.Allocatable).Add(nodeInfo.OversubscriptionResource)
		nodeInfo.Capacity = NewResource(node.Status.Capacity).Add(nodeInfo.OversubscriptionResource)
	}
	nodeInfo.setNodeOthersResource(node)
	nodeInfo.setNodeState(node)
	nodeInfo.setRevocableZone(node)

	return nodeInfo
}
