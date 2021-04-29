package cpumanager

import (
	"reflect"
	"testing"

	"k8s-lx1036/k8s/kubelet/pkg/cm/cpumanager/topology"
	"k8s-lx1036/k8s/kubelet/pkg/cm/cpuset"
)

func TestCPUAccumulatorFreeSockets(t *testing.T) {
	testCases := []struct {
		description   string
		topo          *topology.CPUTopology
		availableCPUs cpuset.CPUSet
		expect        []int
	}{
		{
			"single socket HT, 1 socket free",
			topoSingleSocketHT,
			cpuset.NewCPUSet(0, 1, 2, 3, 4, 5, 6, 7),
			[]int{0},
		},
		{
			"single socket HT, 0 sockets free",
			topoSingleSocketHT,
			cpuset.NewCPUSet(1, 2, 3, 4, 5, 6, 7), // cpu0 已经被分配，所以 socket0 并不是 free 的
			[]int{},
		},
		{
			"dual socket HT, 2 sockets free",
			topoDualSocketHT,
			cpuset.NewCPUSet(0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11), // socket 0/1 都没有逻辑核被分配出去
			[]int{0, 1},
		},
		{
			"dual socket HT, 1 socket free",
			topoDualSocketHT,
			cpuset.NewCPUSet(0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 11), // socket 1 是free的，没有任何逻辑核被分配出去
			[]int{1},
		},
		{
			"dual socket HT, 0 sockets free",
			topoDualSocketHT,
			cpuset.NewCPUSet(0, 2, 3, 4, 5, 6, 7, 8, 9, 11), // socket 0/1 都有逻辑核被分配出去
			[]int{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			acc := newCPUAccumulator(tc.topo, tc.availableCPUs, 0)
			result := acc.freeSockets()
			if !reflect.DeepEqual(result, tc.expect) {
				t.Errorf("[%s] expected %v to equal %v", tc.description, result, tc.expect)
			}
		})
	}
}

func TestCPUAccumulatorFreeCores(t *testing.T) {
	testCases := []struct {
		description   string
		topo          *topology.CPUTopology
		availableCPUs cpuset.CPUSet
		expect        []int
	}{
		{
			"single socket HT, 4 cores free",
			topoSingleSocketHT,
			cpuset.NewCPUSet(0, 1, 2, 3, 4, 5, 6, 7),
			[]int{0, 1, 2, 3},
		},
		{
			"single socket HT, 3 cores free",
			topoSingleSocketHT,
			cpuset.NewCPUSet(0, 1, 2, 4, 5, 6), // core3的逻辑核cpu3/cpu7已经分配出去了，所有只有core0/core1/core2可用
			[]int{0, 1, 2},
		},
		{
			"single socket HT, 0 cores free",
			topoSingleSocketHT,
			cpuset.NewCPUSet(),
			[]int{},
		},
		{
			"single socket HT, 0 cores free (4 partially consumed)",
			topoSingleSocketHT,
			cpuset.NewCPUSet(0, 1, 2, 3),
			[]int{},
		},
		{
			"dual socket HT, 6 cores free",
			topoDualSocketHT,
			cpuset.NewCPUSet(0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11),
			[]int{0, 2, 4, 1, 3, 5},
		},
		{
			"dual socket HT, 5 cores free (1 consumed from socket 0)",
			topoDualSocketHT,
			cpuset.NewCPUSet(2, 1, 3, 4, 5, 7, 8, 9, 10, 11),
			[]int{2, 4, 1, 3, 5}, // 排序规则：先 SocketID, 再 CoreID，最后 ProcessorID
		},
		{
			"dual socket HT, 4 cores free (1 consumed from each socket)",
			topoDualSocketHT,
			cpuset.NewCPUSet(2, 3, 4, 5, 8, 9, 10, 11),
			[]int{2, 4, 3, 5},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			acc := newCPUAccumulator(tc.topo, tc.availableCPUs, 0)
			result := acc.freeCores()
			if !reflect.DeepEqual(result, tc.expect) {
				t.Errorf("[%s] expected %v to equal %v", tc.description, result, tc.expect)
			}
		})
	}
}

func TestCPUAccumulatorFreeCPUs(t *testing.T) {
	testCases := []struct {
		description   string
		topo          *topology.CPUTopology
		availableCPUs cpuset.CPUSet
		expect        []int
	}{
		{
			"single socket HT, 8 cpus free",
			topoSingleSocketHT,
			cpuset.NewCPUSet(0, 1, 2, 3, 4, 5, 6, 7),
			[]int{0, 4, 1, 5, 2, 6, 3, 7},
		},
		{
			"single socket HT, 5 cpus free",
			topoSingleSocketHT,
			cpuset.NewCPUSet(3, 4, 5, 6, 7),
			[]int{4, 5, 6, 3, 7},
		},
		{
			"dual socket HT, 12 cpus free",
			topoDualSocketHT,
			cpuset.NewCPUSet(0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11),
			[]int{0, 6, 2, 8, 4, 10, 1, 7, 3, 9, 5, 11},
		},
		{
			"dual socket HT, 11 cpus free",
			topoDualSocketHT,
			cpuset.NewCPUSet(1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11),
			[]int{6, 2, 8, 4, 10, 1, 7, 3, 9, 5, 11},
		},
		{
			"dual socket HT, 10 cpus free",
			topoDualSocketHT,
			cpuset.NewCPUSet(1, 2, 3, 4, 5, 7, 8, 9, 10, 11),
			[]int{2, 8, 4, 10, 1, 7, 3, 9, 5, 11},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			acc := newCPUAccumulator(tc.topo, tc.availableCPUs, 0)
			result := acc.freeCPUs()
			if !reflect.DeepEqual(result, tc.expect) {
				t.Errorf("[%s] expected %v to equal %v", tc.description, result, tc.expect)
			}
		})
	}
}

func TestCPUAccumulatorTake(t *testing.T) {
	testCases := []struct {
		description     string
		topo            *topology.CPUTopology
		availableCPUs   cpuset.CPUSet
		takeCPUs        []cpuset.CPUSet
		numCPUs         int
		expectSatisfied bool
		expectFailed    bool
	}{
		{
			"take 0 cpus from a single socket HT, require 1",
			topoSingleSocketHT,
			cpuset.NewCPUSet(0, 1, 2, 3, 4, 5, 6, 7),
			[]cpuset.CPUSet{cpuset.NewCPUSet()},
			1,
			false,
			false,
		},
		{
			"take 0 cpus from a single socket HT, require 1, none available",
			topoSingleSocketHT,
			cpuset.NewCPUSet(),
			[]cpuset.CPUSet{cpuset.NewCPUSet()},
			1,
			false,
			true,
		},
		{
			"take 1 cpu from a single socket HT, require 1",
			topoSingleSocketHT,
			cpuset.NewCPUSet(0, 1, 2, 3, 4, 5, 6, 7),
			[]cpuset.CPUSet{cpuset.NewCPUSet(0)},
			1,
			true,
			false,
		},
		{
			"take 1 cpu from a single socket HT, require 2",
			topoSingleSocketHT,
			cpuset.NewCPUSet(0, 1, 2, 3, 4, 5, 6, 7),
			[]cpuset.CPUSet{cpuset.NewCPUSet(0)},
			2,
			false,
			false,
		},
		{
			"take 2 cpu from a single socket HT, require 4, expect failed",
			topoSingleSocketHT,
			cpuset.NewCPUSet(0, 1, 2),
			[]cpuset.CPUSet{cpuset.NewCPUSet(0), cpuset.NewCPUSet(1)},
			4,
			false,
			true,
		},
		{
			"take all cpus one at a time from a single socket HT, require 8",
			topoSingleSocketHT,
			cpuset.NewCPUSet(0, 1, 2, 3, 4, 5, 6, 7),
			[]cpuset.CPUSet{
				cpuset.NewCPUSet(0),
				cpuset.NewCPUSet(1),
				cpuset.NewCPUSet(2),
				cpuset.NewCPUSet(3),
				cpuset.NewCPUSet(4),
				cpuset.NewCPUSet(5),
				cpuset.NewCPUSet(6),
				cpuset.NewCPUSet(7),
			},
			8,
			true,
			false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			acc := newCPUAccumulator(tc.topo, tc.availableCPUs, tc.numCPUs)
			totalTaken := 0
			for _, cpus := range tc.takeCPUs {
				acc.take(cpus)
				totalTaken += cpus.Size()
			}
			if tc.expectSatisfied != acc.isSatisfied() {
				t.Errorf("[%s] expected acc.isSatisfied() to be %t", tc.description, tc.expectSatisfied)
			}
			if tc.expectFailed != acc.isFailed() {
				t.Errorf("[%s] expected acc.isFailed() to be %t", tc.description, tc.expectFailed)
			}

			for _, cpus := range tc.takeCPUs {
				availableCPUs := acc.details.CPUs()
				if cpus.Intersection(availableCPUs).Size() > 0 {
					t.Errorf("[%s] expected intersection of taken cpus [%s] and acc.details.CPUs() [%s] to be empty", tc.description, cpus, availableCPUs)
				}
				if !cpus.IsSubsetOf(acc.result) {
					t.Errorf("[%s] expected [%s] to be a subset of acc.result [%s]", tc.description, cpus, acc.result)
				}
			}

			expNumCPUsNeeded := tc.numCPUs - totalTaken
			if acc.numCPUsNeeded != expNumCPUsNeeded {
				t.Errorf("[%s] expected acc.numCPUsNeeded to be %d (got %d)", tc.description, expNumCPUsNeeded, acc.numCPUsNeeded)
			}
		})
	}
}

// INFO: 根据 topo 来分配 cpu
func TestTakeByTopology(t *testing.T) {
	testCases := []struct {
		description   string
		topo          *topology.CPUTopology
		availableCPUs cpuset.CPUSet
		numCPUs       int
		expErr        string
		expResult     cpuset.CPUSet
	}{
		{
			"take more cpus than are available from single socket with HT",
			topoSingleSocketHT,
			cpuset.NewCPUSet(0, 2, 4, 6),
			5,
			"not enough cpus available to satisfy request",
			cpuset.NewCPUSet(),
		},
		{
			"take zero cpus from single socket with HT",
			topoSingleSocketHT,
			cpuset.NewCPUSet(0, 1, 2, 3, 4, 5, 6, 7),
			0,
			"",
			cpuset.NewCPUSet(),
		},
		{
			"take one cpu from single socket with HT",
			topoSingleSocketHT,
			cpuset.NewCPUSet(0, 1, 2, 3, 4, 5, 6, 7),
			1,
			"",
			cpuset.NewCPUSet(0),
		},
		{
			"take one cpu from single socket with HT, some cpus are taken",
			topoSingleSocketHT,
			cpuset.NewCPUSet(1, 3, 5, 6, 7),
			1,
			"",
			cpuset.NewCPUSet(6), // CoreID:2 有一个逻辑核分配出去，CoreID:1 两个逻辑核都没有分配出去，所以这里应该分配 CoreID:2 的逻辑核
		},
		{
			"take two cpus from single socket with HT",
			topoSingleSocketHT,
			cpuset.NewCPUSet(0, 1, 2, 3, 4, 5, 6, 7),
			2,
			"",
			cpuset.NewCPUSet(0, 4),
		},
		{
			"take all cpus from single socket with HT",
			topoSingleSocketHT,
			cpuset.NewCPUSet(0, 1, 2, 3, 4, 5, 6, 7),
			8,
			"",
			cpuset.NewCPUSet(0, 1, 2, 3, 4, 5, 6, 7),
		},
		{
			"take two cpus from single socket with HT, only one core totally free",
			topoSingleSocketHT,
			cpuset.NewCPUSet(0, 1, 2, 3, 6),
			2,
			"",
			// 这里CoreID:2是一个完整的物理核，需求是2个逻辑核，应该分配这个完整的物理核，而不是从其他的不完整物理核中去分配逻辑核
			cpuset.NewCPUSet(2, 6),
		},
		{
			"take one cpu from dual socket with HT - core from Socket 0",
			topoDualSocketHT,
			cpuset.NewCPUSet(1, 2, 3, 4, 5, 7, 8, 9, 10, 11),
			1,
			"",
			cpuset.NewCPUSet(2),
		},
		{
			"take a socket of cpus from dual socket with HT",
			topoDualSocketHT,
			cpuset.NewCPUSet(0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11),
			6,
			"",
			cpuset.NewCPUSet(0, 2, 4, 6, 8, 10),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			result, err := takeByTopology(tc.topo, tc.availableCPUs, tc.numCPUs)
			if err != nil {
				if err.Error() != tc.expErr {
					t.Errorf("expected error to be [%v] but it was [%v] in test %s", tc.expErr, err, tc.description)
				}

				return
			}

			if !result.Equals(tc.expResult) {
				t.Errorf("expected result [%s] to equal [%s] in test %s", result, tc.expResult, tc.description)
			}
		})
	}
}
