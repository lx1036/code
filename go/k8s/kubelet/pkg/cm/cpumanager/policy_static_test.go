package cpumanager

import (
	"fmt"
	"reflect"
	"testing"

	"k8s-lx1036/k8s/kubelet/pkg/cm/cpumanager/state"
	"k8s-lx1036/k8s/kubelet/pkg/cm/cpumanager/topology"
	"k8s-lx1036/k8s/kubelet/pkg/cm/cpuset"
	"k8s-lx1036/k8s/kubelet/pkg/cm/topologymanager"

	v1 "k8s.io/api/core/v1"
)

var (
	topoDualSocketHT = &topology.CPUTopology{
		NumCPUs:    12, // 每个core就1个processor，说明没有开启超线程
		NumSockets: 2, // 机器上2个cpu，或者是 2 个numa node
		NumCores:   6, // 每个cpu 6个core
		CPUDetails: map[int]topology.CPUInfo{
			0:  {CoreID: 0, SocketID: 0, NUMANodeID: 0},
			1:  {CoreID: 1, SocketID: 1, NUMANodeID: 1},
			2:  {CoreID: 2, SocketID: 0, NUMANodeID: 0},
			3:  {CoreID: 3, SocketID: 1, NUMANodeID: 1},
			4:  {CoreID: 4, SocketID: 0, NUMANodeID: 0},
			5:  {CoreID: 5, SocketID: 1, NUMANodeID: 1},
			6:  {CoreID: 0, SocketID: 0, NUMANodeID: 0},
			7:  {CoreID: 1, SocketID: 1, NUMANodeID: 1},
			8:  {CoreID: 2, SocketID: 0, NUMANodeID: 0},
			9:  {CoreID: 3, SocketID: 1, NUMANodeID: 1},
			10: {CoreID: 4, SocketID: 0, NUMANodeID: 0},
			11: {CoreID: 5, SocketID: 1, NUMANodeID: 1},
		},
	}

	topoSingleSocketHT = &topology.CPUTopology{
		NumCPUs:    8, // 每个core就2个processor，说明开启了超线程
		NumSockets: 1, // 机器上1个cpu，或者是 1 个numa node
		NumCores:   4, // 每个cpu 4个core
		CPUDetails: map[int]topology.CPUInfo{
			0: {CoreID: 0, SocketID: 0, NUMANodeID: 0},
			1: {CoreID: 1, SocketID: 0, NUMANodeID: 0},
			2: {CoreID: 2, SocketID: 0, NUMANodeID: 0},
			3: {CoreID: 3, SocketID: 0, NUMANodeID: 0},
			4: {CoreID: 0, SocketID: 0, NUMANodeID: 0},
			5: {CoreID: 1, SocketID: 0, NUMANodeID: 0},
			6: {CoreID: 2, SocketID: 0, NUMANodeID: 0},
			7: {CoreID: 3, SocketID: 0, NUMANodeID: 0},
		},
	}
)

type staticPolicyTest struct {
	description     string
	topo            *topology.CPUTopology
	numReservedCPUs int
	podUID          string
	containerName   string
	stAssignments   state.ContainerCPUAssignments
	stDefaultCPUSet cpuset.CPUSet
	pod             *v1.Pod
	expErr          error
	expCPUAlloc     bool
	expCSet         cpuset.CPUSet
}

// INFO: 这里没有实现根据topo分配cpu(cpu_assignment.go)逻辑，所以只有第一个测试能通过
func TestStaticPolicyStart(t *testing.T) {
	testCases := []staticPolicyTest{
		{
			description: "non-corrupted state",
			topo:        topoDualSocketHT,
			stAssignments: state.ContainerCPUAssignments{
				"fakePod": map[string]cpuset.CPUSet{
					"0": cpuset.NewCPUSet(0),
				},
			},
			stDefaultCPUSet: cpuset.NewCPUSet(1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11),
			expCSet:         cpuset.NewCPUSet(1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11),
		},
		{
			description:     "empty cpuset",
			topo:            topoDualSocketHT,
			numReservedCPUs: 1,
			stAssignments:   state.ContainerCPUAssignments{},
			stDefaultCPUSet: cpuset.NewCPUSet(),
			expCSet:         cpuset.NewCPUSet(0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11),
		},
		{
			// 因为 cpu0 和 cpu6 在同一个 core 里(CoreID=0)
			description:     "reserved cores 0 & 6 are not present in available cpuset",
			topo:            topoDualSocketHT,
			numReservedCPUs: 2,
			stAssignments:   state.ContainerCPUAssignments{},
			stDefaultCPUSet: cpuset.NewCPUSet(0, 1),
			expErr:          fmt.Errorf("not all reserved cpus: 0,6 are present in defaultCpuSet: 0-1"),
		},
		{
			description: "assigned core 2 is still present in available cpuset",
			topo:        topoDualSocketHT,
			stAssignments: state.ContainerCPUAssignments{
				"fakePod": map[string]cpuset.CPUSet{
					"0": cpuset.NewCPUSet(0, 1, 2),
				},
			},
			stDefaultCPUSet: cpuset.NewCPUSet(2, 3, 4, 5, 6, 7, 8, 9, 10, 11),
			expErr:          fmt.Errorf("pod: fakePod, container: 0 cpuset: 0-2 overlaps with default cpuset 2-11"),
		},
		{
			description: "core 12 is not present in topology but is in state cpuset",
			topo:        topoDualSocketHT,
			stAssignments: state.ContainerCPUAssignments{
				"fakePod": map[string]cpuset.CPUSet{
					"0": cpuset.NewCPUSet(0, 1, 2),
					"1": cpuset.NewCPUSet(3, 4),
				},
			},
			stDefaultCPUSet: cpuset.NewCPUSet(5, 6, 7, 8, 9, 10, 11, 12),
			expErr:          fmt.Errorf("current set of available CPUs 0-11 doesn't match with CPUs in state 0-12"),
		},
		{
			description: "core 11 is present in topology but is not in state cpuset",
			topo:        topoDualSocketHT,
			stAssignments: state.ContainerCPUAssignments{
				"fakePod": map[string]cpuset.CPUSet{
					"0": cpuset.NewCPUSet(0, 1, 2),
					"1": cpuset.NewCPUSet(3, 4),
				},
			},
			stDefaultCPUSet: cpuset.NewCPUSet(5, 6, 7, 8, 9, 10),
			expErr:          fmt.Errorf("current set of available CPUs 0-11 doesn't match with CPUs in state 0-10"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			p, _ := NewStaticPolicy(testCase.topo, testCase.numReservedCPUs,
				cpuset.NewCPUSet(), topologymanager.NewFakeManager())
			policy := p.(*staticPolicy)
			st := &mockState{
				assignments:   testCase.stAssignments,
				defaultCPUSet: testCase.stDefaultCPUSet,
			}
			err := policy.Start(st)
			if !reflect.DeepEqual(err, testCase.expErr) {
				t.Errorf("StaticPolicy Start() error (%v). expected error: %v but got: %v",
					testCase.description, testCase.expErr, err)
			}

			if !testCase.stDefaultCPUSet.IsEmpty() {
				// 不包含 cpuid=0
				for cpuid := 1; cpuid < policy.topology.NumCPUs; cpuid++ {
					if !st.defaultCPUSet.Contains(cpuid) {
						t.Errorf("StaticPolicy Start() error. expected cpuid %d to be present in defaultCPUSet", cpuid)
					}
				}
			}
			if !st.GetDefaultCPUSet().Equals(testCase.expCSet) {
				t.Errorf("State CPUSet is different than expected. Have %q wants: %q", st.GetDefaultCPUSet(),
					testCase.expCSet)
			}
		})
	}
}

func TestStaticPolicyAllocate(test *testing.T) {
	testCases := []staticPolicyTest{
		{
			// INFO: 总共 0-7 八个cpu, 还需要reserved 1个cpu, 要分配出"8000m"，所以报错
			description:     "GuPodSingleCore, SingleSocketHT, ExpectError",
			topo:            topoSingleSocketHT,
			numReservedCPUs: 1,
			stAssignments:   state.ContainerCPUAssignments{},
			stDefaultCPUSet: cpuset.NewCPUSet(0, 1, 2, 3, 4, 5, 6, 7),
			pod:             makePod("fakePod", "fakeContainer2", "8000m", "8000m"),
			expErr:          fmt.Errorf("not enough cpus available to satisfy request"),
			expCPUAlloc:     false,
			expCSet:         cpuset.NewCPUSet(),
		},
	}

	for _, testCase := range testCases {
		test.Run(testCase.description, func(t *testing.T) {
			policy, err := NewStaticPolicy(testCase.topo, testCase.numReservedCPUs,
				cpuset.NewCPUSet(0), topologymanager.NewFakeManager())
			if err != nil {
				panic(err)
			}

			mState := &mockState{
				assignments:   testCase.stAssignments,
				defaultCPUSet: testCase.stDefaultCPUSet,
			}

			container := &testCase.pod.Spec.Containers[0]
			err = policy.Allocate(mState, testCase.pod, container)
			if !reflect.DeepEqual(err, testCase.expErr) {
				t.Errorf("StaticPolicy Allocate() error (%v). expected add error: %v but got: %v",
					testCase.description, testCase.expErr, err)
			}

			if testCase.expCPUAlloc {
				cset, found := mState.assignments[string(testCase.pod.UID)][container.Name]
				if !found {
					t.Errorf("StaticPolicy Allocate() error (%v). expected container %v to be present in assignments %v",
						testCase.description, container.Name, mState.assignments)
				}

				if !reflect.DeepEqual(cset, testCase.expCSet) {
					t.Errorf("StaticPolicy Allocate() error (%v). expected cpuset %v but got %v",
						testCase.description, testCase.expCSet, cset)
				}

				if !cset.Intersection(mState.defaultCPUSet).IsEmpty() {
					t.Errorf("StaticPolicy Allocate() error (%v). expected cpuset %v to be disoint from the shared cpuset %v",
						testCase.description, cset, mState.defaultCPUSet)
				}
			}

			if !testCase.expCPUAlloc {
				_, found := mState.assignments[string(testCase.pod.UID)][container.Name]
				if found {
					t.Errorf("StaticPolicy Allocate() error (%v). Did not expect container %v to be present in assignments %v",
						testCase.description, container.Name, mState.assignments)
				}
			}
		})
	}
}
