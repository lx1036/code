package noderesources

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"k8s-lx1036/k8s/scheduler/pkg/scheduler/apis/config"
	framework "k8s-lx1036/k8s/scheduler/pkg/scheduler/framework"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	v1helper "k8s.io/kubernetes/pkg/apis/core/v1/helper"
)

var (
	extendedResourceA = v1.ResourceName("example.com/aaa")
	hugePageResourceA = v1helper.HugePageResourceName(resource.MustParse("2Mi"))
	extendedResourceB = v1.ResourceName("example.com/bbb")
)

func TestEnoughRequests(test *testing.T) {
	enoughPodsTests := []struct {
		pod                       *v1.Pod
		nodeInfo                  *framework.NodeInfo
		name                      string
		args                      config.NodeResourcesFitArgs
		wantInsufficientResources []InsufficientResource
		wantStatus                *framework.Status
	}{
		{
			pod:                       &v1.Pod{},
			nodeInfo:                  framework.NewNodeInfo(newResourcePod(framework.Resource{MilliCPU: 10, Memory: 20})),
			name:                      "no resources requested always fits",
			wantInsufficientResources: []InsufficientResource{},
		},
		{
			pod: newResourcePod(framework.Resource{MilliCPU: 1, Memory: 1}),
			// 已经有了一个 cpu:10 memory:20 的 pod
			nodeInfo:   framework.NewNodeInfo(newResourcePod(framework.Resource{MilliCPU: 10, Memory: 20})),
			name:       "too many resources fails",
			wantStatus: framework.NewStatus(framework.Unschedulable, getErrReason(v1.ResourceCPU), getErrReason(v1.ResourceMemory)),
			wantInsufficientResources: []InsufficientResource{
				{v1.ResourceCPU, getErrReason(v1.ResourceCPU), 1, 10, 10},
				{v1.ResourceMemory, getErrReason(v1.ResourceMemory), 1, 20, 20},
			},
		},
		{
			// cpu: 3, memory: 1
			pod:        newResourceInitPod(newResourcePod(framework.Resource{MilliCPU: 1, Memory: 1}), framework.Resource{MilliCPU: 3, Memory: 1}),
			nodeInfo:   framework.NewNodeInfo(newResourcePod(framework.Resource{MilliCPU: 8, Memory: 19})),
			name:       "too many resources fails due to init container cpu",
			wantStatus: framework.NewStatus(framework.Unschedulable, getErrReason(v1.ResourceCPU)),
			wantInsufficientResources: []InsufficientResource{
				{v1.ResourceCPU, getErrReason(v1.ResourceCPU), 3, 8, 10},
			},
		},
		{
			pod:                       newResourceInitPod(newResourcePod(framework.Resource{MilliCPU: 1, Memory: 1}), framework.Resource{MilliCPU: 1, Memory: 1}),
			nodeInfo:                  framework.NewNodeInfo(newResourcePod(framework.Resource{MilliCPU: 9, Memory: 19})),
			name:                      "init container fits because it's the max, not sum, of containers and init containers",
			wantInsufficientResources: []InsufficientResource{},
		},
		{
			pod:                       newResourceInitPod(newResourcePod(framework.Resource{MilliCPU: 4, Memory: 1}), framework.Resource{MilliCPU: 5, Memory: 1}),
			nodeInfo:                  framework.NewNodeInfo(newResourcePod(framework.Resource{MilliCPU: 5, Memory: 19})),
			name:                      "equal edge case for init container",
			wantInsufficientResources: []InsufficientResource{},
		},
		{
			pod:      newResourcePod(framework.Resource{MilliCPU: 1, Memory: 1, ScalarResources: map[v1.ResourceName]int64{extendedResourceB: 1}}),
			nodeInfo: framework.NewNodeInfo(newResourcePod(framework.Resource{MilliCPU: 0, Memory: 0})),
			args: config.NodeResourcesFitArgs{
				IgnoredResources: []string{"example.com/bbb"},
			},
			name:                      "skip checking ignored extended resource",
			wantInsufficientResources: []InsufficientResource{},
		},
		{
			pod: newResourceOverheadPod(
				newResourcePod(framework.Resource{MilliCPU: 1, Memory: 1}),
				v1.ResourceList{v1.ResourceCPU: resource.MustParse("3m"), v1.ResourceMemory: resource.MustParse("13")},
			),
			nodeInfo:                  framework.NewNodeInfo(newResourcePod(framework.Resource{MilliCPU: 5, Memory: 5})),
			name:                      "resources + pod overhead fits",
			wantInsufficientResources: []InsufficientResource{},
		},
	}

	for _, fixture := range enoughPodsTests {
		test.Run(fixture.name, func(t *testing.T) {
			node := v1.Node{
				Status: v1.NodeStatus{ // 10m, 20B
					Capacity:    makeResources(10, 20, 32, 5, 20, 5).Capacity,
					Allocatable: makeAllocatableResources(10, 20, 32, 5, 20, 5),
				},
			}
			fixture.nodeInfo.SetNode(&node)

			plugin, err := NewFit(&fixture.args, nil)
			if err != nil {
				t.Fatal(err)
			}
			cycleState := framework.NewCycleState()
			preFilterStatus := plugin.(framework.PreFilterPlugin).PreFilter(context.Background(), cycleState, fixture.pod)
			if !preFilterStatus.IsSuccess() {
				t.Errorf("prefilter failed with status: %v", preFilterStatus)
			}

			gotStatus := plugin.(framework.FilterPlugin).Filter(context.Background(), cycleState, fixture.pod, fixture.nodeInfo)
			if !reflect.DeepEqual(gotStatus, fixture.wantStatus) {
				t.Errorf("status does not match: %v, want: %v", gotStatus, fixture.wantStatus)
			}

			gotInsufficientResources := fitsRequest(computePodResourceRequest(fixture.pod),
				fixture.nodeInfo, plugin.(*Fit).ignoredResources, plugin.(*Fit).ignoredResourceGroups)
			if !reflect.DeepEqual(gotInsufficientResources, fixture.wantInsufficientResources) {
				t.Errorf("insufficient resources do not match: %+v, want: %v", gotInsufficientResources, fixture.wantInsufficientResources)
			}
		})
	}
}

func makeResources(milliCPU, memory, pods, extendedA, storage, hugePageA int64) v1.NodeResources {
	return v1.NodeResources{
		Capacity: v1.ResourceList{
			v1.ResourceCPU:              *resource.NewMilliQuantity(milliCPU, resource.DecimalSI),
			v1.ResourceMemory:           *resource.NewQuantity(memory, resource.BinarySI),
			v1.ResourcePods:             *resource.NewQuantity(pods, resource.DecimalSI),
			extendedResourceA:           *resource.NewQuantity(extendedA, resource.DecimalSI),
			v1.ResourceEphemeralStorage: *resource.NewQuantity(storage, resource.BinarySI),
			hugePageResourceA:           *resource.NewQuantity(hugePageA, resource.BinarySI),
		},
	}
}

func makeAllocatableResources(milliCPU, memory, pods, extendedA, storage, hugePageA int64) v1.ResourceList {
	return v1.ResourceList{
		v1.ResourceCPU:              *resource.NewMilliQuantity(milliCPU, resource.DecimalSI),
		v1.ResourceMemory:           *resource.NewQuantity(memory, resource.BinarySI),
		v1.ResourcePods:             *resource.NewQuantity(pods, resource.DecimalSI),
		extendedResourceA:           *resource.NewQuantity(extendedA, resource.DecimalSI),
		v1.ResourceEphemeralStorage: *resource.NewQuantity(storage, resource.BinarySI),
		hugePageResourceA:           *resource.NewQuantity(hugePageA, resource.BinarySI),
	}
}

func newResourcePod(usage ...framework.Resource) *v1.Pod {
	var containers []v1.Container
	for _, req := range usage {
		containers = append(containers, v1.Container{
			Resources: v1.ResourceRequirements{Requests: req.ResourceList()},
		})
	}
	return &v1.Pod{
		Spec: v1.PodSpec{
			Containers: containers,
		},
	}
}

func getErrReason(rn v1.ResourceName) string {
	return fmt.Sprintf("Insufficient %v", rn)
}

func newResourceInitPod(pod *v1.Pod, usage ...framework.Resource) *v1.Pod {
	pod.Spec.InitContainers = newResourcePod(usage...).Spec.Containers
	return pod
}

func newResourceOverheadPod(pod *v1.Pod, overhead v1.ResourceList) *v1.Pod {
	pod.Spec.Overhead = overhead
	return pod
}
