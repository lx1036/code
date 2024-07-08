package api

import v1 "k8s.io/api/core/v1"

// Resource struct defines all the resource type
type Resource struct {
	MilliCPU float64
	Memory   float64

	// ScalarResources
	ScalarResources map[v1.ResourceName]float64

	// MaxTaskNum is only used by predicates; it should NOT
	// be accounted in other operators, e.g. Add.
	MaxTaskNum int
}
