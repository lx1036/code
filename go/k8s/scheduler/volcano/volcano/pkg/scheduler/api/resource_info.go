package api

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
	v1helper "k8s.io/kubernetes/pkg/scheduler/util"
	"strings"
)

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

// AddScalar adds a resource by a scalar value of this resource.
func (r *Resource) AddScalar(name v1.ResourceName, quantity float64) {
	r.SetScalar(name, r.ScalarResources[name]+quantity)
}

func (r *Resource) SetScalar(name v1.ResourceName, quantity float64) {
	if r.ScalarResources == nil {
		r.ScalarResources = map[v1.ResourceName]float64{}
	}
	r.ScalarResources[name] = quantity
}

// Add is used to add two given resources
func (r *Resource) Add(rr *Resource) *Resource {
	r.MilliCPU += rr.MilliCPU
	r.Memory += rr.Memory

	for rName, rQuant := range rr.ScalarResources {
		if r.ScalarResources == nil {
			r.ScalarResources = map[v1.ResourceName]float64{}
		}
		r.ScalarResources[rName] += rQuant
	}

	return r
}

func EmptyResource() *Resource {
	return &Resource{}
}

// NewResource creates a new resource object from resource list
func NewResource(rl v1.ResourceList) *Resource {
	r := EmptyResource()
	for rName, rQuant := range rl {
		switch rName {
		case v1.ResourceCPU:
			r.MilliCPU += float64(rQuant.MilliValue())
		case v1.ResourceMemory:
			r.Memory += float64(rQuant.Value())
		case v1.ResourcePods:
			r.MaxTaskNum += int(rQuant.Value())
			r.AddScalar(rName, float64(rQuant.Value()))
		case v1.ResourceEphemeralStorage:
			r.AddScalar(rName, float64(rQuant.MilliValue()))
		default:
			if IsCountQuota(rName) {
				continue
			}

			//NOTE: When converting this back to k8s resource, we need record the format as well as / 1000
			if v1helper.IsScalarResourceName(rName) {
				ignore := false
				IgnoredDevicesList.Range(func(_ int, val string) bool {
					if rName.String() == val {
						ignore = true
						return false
					}
					return true
				})
				if !ignore {
					r.AddScalar(rName, float64(rQuant.MilliValue()))
				} else {
					klog.V(4).Infof("Ignoring resource %s", rName.String())
				}
			}
		}
	}

	return r
}

func IsCountQuota(name v1.ResourceName) bool {
	return strings.HasPrefix(string(name), "count/")
}
