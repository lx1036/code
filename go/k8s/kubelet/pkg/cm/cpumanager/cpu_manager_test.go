package cpumanager

import (
	"fmt"
	"testing"

	"k8s-lx1036/k8s/kubelet/pkg/cm/containermap"
	"k8s-lx1036/k8s/kubelet/pkg/cm/cpumanager/state"
	"k8s-lx1036/k8s/kubelet/pkg/cm/cpuset"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

func TestReconcileState(t *testing.T) {
	testCases := []struct {
		description                  string
		activePods                   []*v1.Pod
		podStatus                    v1.PodStatus
		podStatusFound               bool
		stAssignments                state.ContainerCPUAssignments
		stDefaultCPUSet              cpuset.CPUSet
		updateErr                    error
		expectSucceededContainerName string
		expectFailedContainerName    string
	}{
		{
			description: "cpu manager reconclie - no error",
			activePods: []*v1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "fakePodName",
						UID:  "fakePodUID",
					},
					Spec: v1.PodSpec{
						Containers: []v1.Container{
							{
								Name: "fakeContainerName",
							},
						},
					},
				},
			},
			podStatus: v1.PodStatus{
				ContainerStatuses: []v1.ContainerStatus{
					{
						Name:        "fakeContainerName",
						ContainerID: "docker://fakeContainerID",
						State: v1.ContainerState{
							Running: &v1.ContainerStateRunning{},
						},
					},
				},
			},
			podStatusFound: true,
			stAssignments: state.ContainerCPUAssignments{
				"fakePodUID": map[string]cpuset.CPUSet{
					"fakeContainerName": cpuset.NewCPUSet(1, 2),
				},
			},
			stDefaultCPUSet:              cpuset.NewCPUSet(3, 4, 5, 6, 7),
			updateErr:                    nil,
			expectSucceededContainerName: "fakeContainerName",
			expectFailedContainerName:    "",
		},
		{
			description: "cpu manager reconcile init container - no error",
			activePods: []*v1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "fakePodName",
						UID:  "fakePodUID",
					},
					Spec: v1.PodSpec{
						InitContainers: []v1.Container{
							{
								Name: "fakeContainerName",
							},
						},
					},
				},
			},
			podStatus: v1.PodStatus{
				InitContainerStatuses: []v1.ContainerStatus{
					{
						Name:        "fakeContainerName",
						ContainerID: "docker://fakeContainerID",
						State: v1.ContainerState{
							Running: &v1.ContainerStateRunning{},
						},
					},
				},
			},
			podStatusFound: true,
			stAssignments: state.ContainerCPUAssignments{
				"fakePodUID": map[string]cpuset.CPUSet{
					"fakeContainerName": cpuset.NewCPUSet(1, 2),
				},
			},
			stDefaultCPUSet:              cpuset.NewCPUSet(3, 4, 5, 6, 7),
			updateErr:                    nil,
			expectSucceededContainerName: "fakeContainerName",
			expectFailedContainerName:    "",
		},
		{
			description: "cpu manager reconcile - pod status not found",
			activePods: []*v1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "fakePodName",
						UID:  "fakePodUID",
					},
					Spec: v1.PodSpec{
						Containers: []v1.Container{
							{
								Name: "fakeContainerName",
							},
						},
					},
				},
			},
			podStatus:                    v1.PodStatus{},
			podStatusFound:               false,
			stAssignments:                state.ContainerCPUAssignments{},
			stDefaultCPUSet:              cpuset.NewCPUSet(),
			updateErr:                    nil,
			expectSucceededContainerName: "",
			expectFailedContainerName:    "",
		},
		{
			description: "cpu manager reconcile - container state not found",
			activePods: []*v1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "fakePodName",
						UID:  "fakePodUID",
					},
					Spec: v1.PodSpec{
						Containers: []v1.Container{
							{
								Name: "fakeContainerName",
							},
						},
					},
				},
			},
			podStatus: v1.PodStatus{
				ContainerStatuses: []v1.ContainerStatus{
					{
						Name:        "fakeContainerName1", // 不是 "fakeContainerName"
						ContainerID: "docker://fakeContainerID",
					},
				},
			},
			podStatusFound:               true,
			stAssignments:                state.ContainerCPUAssignments{},
			stDefaultCPUSet:              cpuset.NewCPUSet(),
			updateErr:                    nil,
			expectSucceededContainerName: "",
			expectFailedContainerName:    "fakeContainerName",
		},
		{
			description: "cpu manager reconclie - cpuset is empty",
			activePods: []*v1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "fakePodName",
						UID:  "fakePodUID",
					},
					Spec: v1.PodSpec{
						Containers: []v1.Container{
							{
								Name: "fakeContainerName",
							},
						},
					},
				},
			},
			podStatus: v1.PodStatus{
				ContainerStatuses: []v1.ContainerStatus{
					{
						Name:        "fakeContainerName",
						ContainerID: "docker://fakeContainerID",
						State: v1.ContainerState{
							Running: &v1.ContainerStateRunning{},
						},
					},
				},
			},
			podStatusFound: true,
			stAssignments: state.ContainerCPUAssignments{
				"fakePodUID": map[string]cpuset.CPUSet{
					"fakeContainerName": cpuset.NewCPUSet(),
				},
			},
			stDefaultCPUSet:              cpuset.NewCPUSet(1, 2, 3, 4, 5, 6, 7),
			updateErr:                    nil,
			expectSucceededContainerName: "",
			expectFailedContainerName:    "fakeContainerName",
		},
		{
			description: "cpu manager reconclie - container update error",
			activePods: []*v1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "fakePodName",
						UID:  "fakePodUID",
					},
					Spec: v1.PodSpec{
						Containers: []v1.Container{
							{
								Name: "fakeContainerName",
							},
						},
					},
				},
			},
			podStatus: v1.PodStatus{
				ContainerStatuses: []v1.ContainerStatus{
					{
						Name:        "fakeContainerName",
						ContainerID: "docker://fakeContainerID",
						State: v1.ContainerState{
							Running: &v1.ContainerStateRunning{},
						},
					},
				},
			},
			podStatusFound: true,
			stAssignments: state.ContainerCPUAssignments{
				"fakePodUID": map[string]cpuset.CPUSet{
					"fakeContainerName": cpuset.NewCPUSet(1, 2),
				},
			},
			stDefaultCPUSet:              cpuset.NewCPUSet(3, 4, 5, 6, 7),
			updateErr:                    fmt.Errorf("fake container update error"),
			expectSucceededContainerName: "",
			expectFailedContainerName:    "fakeContainerName",
		},
	}

	for _, testCase := range testCases {
		mgr := &manager{
			policy: &mockPolicy{
				err: nil,
			},
			state: &mockState{
				assignments:   testCase.stAssignments,
				defaultCPUSet: testCase.stDefaultCPUSet,
			},
			containerRuntime: mockRuntimeService{
				err: testCase.updateErr,
			},
			containerMap: containermap.NewContainerMap(),
			activePods: func() []*v1.Pod {
				return testCase.activePods
			},
			podStatusProvider: mockPodStatusProvider{
				podStatus: testCase.podStatus,
				found:     testCase.podStatusFound,
			},
			sourcesReady: &sourcesReadyStub{},
		}

		success, failure := mgr.reconcileState()
		klog.InfoS("reconcileState", "success", success, "failure", failure)

		if testCase.expectSucceededContainerName != "" {
			// Search succeeded reconciled containers for the supplied name.
			foundSucceededContainer := false
			for _, reconciled := range success {
				if reconciled.containerName == testCase.expectSucceededContainerName {
					foundSucceededContainer = true
					break
				}
			}
			if !foundSucceededContainer {
				t.Errorf("%v", testCase.description)
				t.Errorf("Expected reconciliation success for container: %s", testCase.expectSucceededContainerName)
			}
		}

		if testCase.expectFailedContainerName != "" {
			// Search failed reconciled containers for the supplied name.
			foundFailedContainer := false
			for _, reconciled := range failure {
				if reconciled.containerName == testCase.expectFailedContainerName {
					foundFailedContainer = true
					break
				}
			}
			if !foundFailedContainer {
				t.Errorf("%v", testCase.description)
				t.Errorf("Expected reconciliation failure for container: %s", testCase.expectFailedContainerName)
			}
		}
	}
}
