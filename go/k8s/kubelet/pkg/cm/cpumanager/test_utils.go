package cpumanager

import (
	"k8s-lx1036/k8s/kubelet/pkg/cm/cpumanager/state"
	"k8s-lx1036/k8s/kubelet/pkg/cm/cpuset"
	"k8s-lx1036/k8s/kubelet/pkg/cm/topologymanager"
	"k8s.io/apimachinery/pkg/api/resource"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	runtimeapi "k8s.io/cri-api/pkg/apis/runtime/v1alpha2"
)

type mockPolicy struct {
	err error
}

func (p *mockPolicy) Name() string {
	return "mock"
}

func (p *mockPolicy) Start(s state.State) error {
	return p.err
}

func (p *mockPolicy) Allocate(s state.State, pod *v1.Pod, container *v1.Container) error {
	return p.err
}

func (p *mockPolicy) RemoveContainer(s state.State, podUID string, containerName string) error {
	return p.err
}

func (p *mockPolicy) GetTopologyHints(s state.State, pod *v1.Pod, container *v1.Container) map[string][]topologymanager.TopologyHint {
	return nil
}

type mockState struct {
	assignments   state.ContainerCPUAssignments
	defaultCPUSet cpuset.CPUSet
}

func (s *mockState) GetCPUSet(podUID string, containerName string) (cpuset.CPUSet, bool) {
	res, ok := s.assignments[podUID][containerName]
	return res.Clone(), ok
}

func (s *mockState) GetDefaultCPUSet() cpuset.CPUSet {
	return s.defaultCPUSet.Clone()
}

func (s *mockState) GetCPUSetOrDefault(podUID string, containerName string) cpuset.CPUSet {
	if res, ok := s.GetCPUSet(podUID, containerName); ok {
		return res
	}
	return s.GetDefaultCPUSet()
}

func (s *mockState) SetCPUSet(podUID string, containerName string, cset cpuset.CPUSet) {
	if _, exists := s.assignments[podUID]; !exists {
		s.assignments[podUID] = make(map[string]cpuset.CPUSet)
	}
	s.assignments[podUID][containerName] = cset
}

func (s *mockState) SetDefaultCPUSet(cset cpuset.CPUSet) {
	s.defaultCPUSet = cset
}

func (s *mockState) Delete(podUID string, containerName string) {
	delete(s.assignments[podUID], containerName)
	if len(s.assignments[podUID]) == 0 {
		delete(s.assignments, podUID)
	}
}

func (s *mockState) ClearState() {
	s.defaultCPUSet = cpuset.CPUSet{}
	s.assignments = make(state.ContainerCPUAssignments)
}

func (s *mockState) SetCPUAssignments(a state.ContainerCPUAssignments) {
	s.assignments = a.Clone()
}

func (s *mockState) GetCPUAssignments() state.ContainerCPUAssignments {
	return s.assignments.Clone()
}

type mockRuntimeService struct {
	err error
}

func (rt mockRuntimeService) UpdateContainerResources(id string, resources *runtimeapi.LinuxContainerResources) error {
	return rt.err
}

type mockPodStatusProvider struct {
	podStatus v1.PodStatus
	found     bool
}

func (podStatusProvider mockPodStatusProvider) GetPodStatus(uid types.UID) (v1.PodStatus, bool) {
	return podStatusProvider.podStatus, podStatusProvider.found
}

func makePod(podUID, containerName, cpuRequest, cpuLimit string) *v1.Pod {
	pod := &v1.Pod{
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Resources: v1.ResourceRequirements{
						Requests: v1.ResourceList{
							v1.ResourceName(v1.ResourceCPU):    resource.MustParse(cpuRequest),
							v1.ResourceName(v1.ResourceMemory): resource.MustParse("1G"),
						},
						Limits: v1.ResourceList{
							v1.ResourceName(v1.ResourceCPU):    resource.MustParse(cpuLimit),
							v1.ResourceName(v1.ResourceMemory): resource.MustParse("1G"),
						},
					},
				},
			},
		},
	}

	pod.UID = types.UID(podUID)
	pod.Spec.Containers[0].Name = containerName

	return pod
}
