package api

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type DeschedulerPolicy struct {
	metav1.TypeMeta

	// Strategies
	Strategies StrategyList

	// NodeSelector for a set of nodes to operate over
	NodeSelector *string

	// EvictLocalStoragePods allows pods using local storage to be evicted.
	EvictLocalStoragePods *bool

	// EvictSystemCriticalPods allows eviction of pods of any priority (including Kubernetes system pods)
	EvictSystemCriticalPods *bool

	// IgnorePVCPods prevents pods with PVCs from being evicted.
	IgnorePVCPods *bool

	// MaxNoOfPodsToEvictPerNode restricts maximum of pods to be evicted per node.
	MaxNoOfPodsToEvictPerNode *int
}

type StrategyName string
type StrategyList map[StrategyName]DeschedulerStrategy

type DeschedulerStrategy struct {
	// Enabled or disabled
	Enabled bool

	// Weight
	Weight int

	// Strategy parameters
	Params *StrategyParameters
}

// Besides Namespaces only one of its members may be specified
type StrategyParameters struct {
	NodeResourceUtilizationThresholds *NodeResourceUtilizationThresholds
	NodeAffinityType                  []string
	PodsHavingTooManyRestarts         *PodsHavingTooManyRestarts
	PodLifeTime                       *PodLifeTime
	RemoveDuplicates                  *RemoveDuplicates
	IncludeSoftConstraints            bool
	Namespaces                        *Namespaces
	ThresholdPriority                 *int32
	ThresholdPriorityClassName        string
	LabelSelector                     *metav1.LabelSelector
}

// Namespaces carries a list of included/excluded namespaces
// for which a given strategy is applicable
type Namespaces struct {
	Include []string
	Exclude []string
}

type Percentage float64
type ResourceThresholds map[v1.ResourceName]Percentage

type NodeResourceUtilizationThresholds struct {
	Thresholds       ResourceThresholds
	TargetThresholds ResourceThresholds
	NumberOfNodes    int
}

type PodsHavingTooManyRestarts struct {
	PodRestartThreshold     int32
	IncludingInitContainers bool
}

type RemoveDuplicates struct {
	ExcludeOwnerKinds []string
}

type PodLifeTime struct {
	MaxPodLifeTimeSeconds *uint
	PodStatusPhases       []string
}
