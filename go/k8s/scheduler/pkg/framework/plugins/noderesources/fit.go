package noderesources

import (
	"context"
	"fmt"
	"strings"

	"k8s-lx1036/k8s/scheduler/pkg/apis/config"
	framework "k8s-lx1036/k8s/scheduler/pkg/framework"

	v1 "k8s.io/api/core/v1"
	metav1validation "k8s.io/apimachinery/pkg/apis/meta/v1/validation"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/validation/field"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
	v1helper "k8s.io/kubernetes/pkg/apis/core/v1/helper"
	"k8s.io/kubernetes/pkg/features"
)

const (
	FitName = "NodeResourcesFit"

	preFilterStateKey = "PreFilter" + FitName
)

// Fit is a plugin that checks if a node has sufficient resources.
type Fit struct {
	ignoredResources      sets.String
	ignoredResourceGroups sets.String
}

func (f *Fit) Name() string {
	return FitName
}

func getFitArgs(obj runtime.Object) (config.NodeResourcesFitArgs, error) {
	ptr, ok := obj.(*config.NodeResourcesFitArgs)
	if !ok {
		return config.NodeResourcesFitArgs{}, fmt.Errorf("want args to be of type NodeResourcesFitArgs, got %T", obj)
	}
	return *ptr, nil
}

func validateFitArgs(args config.NodeResourcesFitArgs) error {
	var allErrs field.ErrorList
	resPath := field.NewPath("ignoredResources")
	for i, res := range args.IgnoredResources {
		path := resPath.Index(i)
		if errs := metav1validation.ValidateLabelName(res, path); len(errs) != 0 {
			allErrs = append(allErrs, errs...)
		}
	}

	groupPath := field.NewPath("ignoredResourceGroups")
	for i, group := range args.IgnoredResourceGroups {
		path := groupPath.Index(i)
		if strings.Contains(group, "/") {
			allErrs = append(allErrs, field.Invalid(path, group, "resource group name can't contain '/'"))
		}
		if errs := metav1validation.ValidateLabelName(group, path); len(errs) != 0 {
			allErrs = append(allErrs, errs...)
		}
	}

	if len(allErrs) == 0 {
		return nil
	}
	return allErrs.ToAggregate()
}

func (f *Fit) PreFilter(ctx context.Context, cycleState *framework.CycleState, pod *v1.Pod) *framework.Status {
	cycleState.Write(preFilterStateKey, computePodResourceRequest(pod))
	return nil
}

func (f *Fit) PreFilterExtensions() framework.PreFilterExtensions {
	return nil
}

func getPreFilterState(cycleState *framework.CycleState) (*preFilterState, error) {
	c, err := cycleState.Read(preFilterStateKey)
	if err != nil {
		// preFilterState doesn't exist, likely PreFilter wasn't invoked.
		return nil, fmt.Errorf("error reading %q from cycleState: %v", preFilterStateKey, err)
	}

	s, ok := c.(*preFilterState)
	if !ok {
		return nil, fmt.Errorf("%+v  convert to NodeResourcesFit.preFilterState error", c)
	}
	return s, nil
}

func (f *Fit) Filter(ctx context.Context, cycleState *framework.CycleState, pod *v1.Pod, nodeInfo *framework.NodeInfo) *framework.Status {
	podResourceRequest, err := getPreFilterState(cycleState)
	if err != nil {
		return framework.NewStatus(framework.Error, err.Error())
	}

	insufficientResources := fitsRequest(podResourceRequest, nodeInfo, f.ignoredResources, f.ignoredResourceGroups)
	if len(insufficientResources) != 0 {
		failureReasons := make([]string, 0, len(insufficientResources))
		for _, r := range insufficientResources {
			failureReasons = append(failureReasons, r.Reason)
		}

		return framework.NewStatus(framework.Unschedulable, failureReasons...)
	}

	return nil
}

func NewFit(plArgs runtime.Object, _ framework.FrameworkHandle) (framework.Plugin, error) {
	args, err := getFitArgs(plArgs)
	if err != nil {
		return nil, err
	}

	if err := validateFitArgs(args); err != nil {
		return nil, err
	}

	return &Fit{
		ignoredResources:      sets.NewString(args.IgnoredResources...),
		ignoredResourceGroups: sets.NewString(args.IgnoredResourceGroups...),
	}, nil
}

type preFilterState struct {
	framework.Resource
}

func (p *preFilterState) Clone() framework.StateData {
	return p
}

// computePodResourceRequest returns a framework.Resource that covers the largest
// width in each resource dimension. Because init-containers run sequentially, we collect
// the max in each dimension iteratively. In contrast, we sum the resource vectors for
// regular containers since they run simultaneously.
//
// If Pod Overhead is specified and the feature gate is set, the resources defined for Overhead
// are added to the calculated Resource request sum
//
// Example:
//
// Pod:
//   InitContainers
//     IC1:
//       CPU: 2
//       Memory: 1G
//     IC2:
//       CPU: 2
//       Memory: 3G
//   Containers
//     C1:
//       CPU: 2
//       Memory: 1G
//     C2:
//       CPU: 1
//       Memory: 1G
//
// Result: CPU: 3, Memory: 3G
func computePodResourceRequest(pod *v1.Pod) *preFilterState {
	result := &preFilterState{}
	for _, container := range pod.Spec.Containers {
		result.Add(container.Resources.Requests)
	}

	// take max_resource(sum_pod, any_init_container)
	for _, container := range pod.Spec.InitContainers {
		result.SetMaxResource(container.Resources.Requests)
	}

	// If Overhead is being utilized, add to the total requests for the pod
	if pod.Spec.Overhead != nil && utilfeature.DefaultFeatureGate.Enabled(features.PodOverhead) {
		result.Add(pod.Spec.Overhead)
	}

	return result
}

// InsufficientResource describes what kind of resource limit is hit and caused the pod to not fit the node.
type InsufficientResource struct {
	ResourceName v1.ResourceName
	// We explicitly have a parameter for reason to avoid formatting a message on the fly
	// for common resources, which is expensive for cluster autoscaler simulations.
	Reason    string
	Requested int64
	Used      int64
	Capacity  int64
}

func fitsRequest(podRequest *preFilterState, nodeInfo *framework.NodeInfo, ignoredExtendedResources,
	ignoredResourceGroups sets.String) []InsufficientResource {
	insufficientResources := make([]InsufficientResource, 0, 4)

	allowedPodNumber := nodeInfo.Allocatable.AllowedPodNumber
	if len(nodeInfo.Pods)+1 > allowedPodNumber {
		insufficientResources = append(insufficientResources, InsufficientResource{
			v1.ResourcePods,
			"Too many pods",
			1,
			int64(len(nodeInfo.Pods)),
			int64(allowedPodNumber),
		})
	}

	if podRequest.MilliCPU == 0 &&
		podRequest.Memory == 0 &&
		podRequest.EphemeralStorage == 0 &&
		len(podRequest.ScalarResources) == 0 {
		return insufficientResources
	}

	// INFO: 如果 (pod request cpu + cpu usage) > node allocatable cpu，则 cpu 不足
	if nodeInfo.Allocatable.MilliCPU < (podRequest.MilliCPU + nodeInfo.Requested.MilliCPU) {
		insufficientResources = append(insufficientResources, InsufficientResource{
			v1.ResourceCPU,
			"Insufficient cpu",
			podRequest.MilliCPU,
			nodeInfo.Requested.MilliCPU,
			nodeInfo.Allocatable.MilliCPU,
		})
	}
	if nodeInfo.Allocatable.Memory < podRequest.Memory+nodeInfo.Requested.Memory {
		insufficientResources = append(insufficientResources, InsufficientResource{
			v1.ResourceMemory,
			"Insufficient memory",
			podRequest.Memory,
			nodeInfo.Requested.Memory,
			nodeInfo.Allocatable.Memory,
		})
	}
	if nodeInfo.Allocatable.EphemeralStorage < podRequest.EphemeralStorage+nodeInfo.Requested.EphemeralStorage {
		insufficientResources = append(insufficientResources, InsufficientResource{
			v1.ResourceEphemeralStorage,
			"Insufficient ephemeral-storage",
			podRequest.EphemeralStorage,
			nodeInfo.Requested.EphemeralStorage,
			nodeInfo.Allocatable.EphemeralStorage,
		})
	}

	// INFO: 自定义资源不足
	for rName, rQuant := range podRequest.ScalarResources {
		if v1helper.IsExtendedResourceName(rName) {
			// If this resource is one of the extended resources that should be ignored, we will skip checking it.
			// rName is guaranteed to have a slash due to API validation.
			var rNamePrefix string
			if ignoredResourceGroups.Len() > 0 {
				rNamePrefix = strings.Split(string(rName), "/")[0]
			}
			if ignoredExtendedResources.Has(string(rName)) || ignoredResourceGroups.Has(rNamePrefix) {
				continue
			}
		}
		if nodeInfo.Allocatable.ScalarResources[rName] < rQuant+nodeInfo.Requested.ScalarResources[rName] {
			insufficientResources = append(insufficientResources, InsufficientResource{
				rName,
				fmt.Sprintf("Insufficient %v", rName),
				podRequest.ScalarResources[rName],
				nodeInfo.Requested.ScalarResources[rName],
				nodeInfo.Allocatable.ScalarResources[rName],
			})
		}
	}

	return insufficientResources
}
