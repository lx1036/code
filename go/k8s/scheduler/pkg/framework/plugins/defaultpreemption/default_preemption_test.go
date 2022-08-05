package defaultpreemption

import (
	internalcache "k8s-lx1036/k8s/scheduler/pkg/internal/cache"
	corev1 "k8s.io/api/core/v1"
	"sort"
	"strings"
	"testing"

	configv1 "k8s-lx1036/k8s/scheduler/pkg/apis/config/v1"
	"k8s-lx1036/k8s/scheduler/pkg/framework"
	schedulertesting "k8s-lx1036/k8s/scheduler/pkg/testing"

	"github.com/google/go-cmp/cmp"
)

var (
	negPriority, lowPriority, midPriority, highPriority, veryHighPriority = int32(-100), int32(0), int32(100), int32(1000), int32(10000)

	veryLargeRes = map[corev1.ResourceName]string{
		corev1.ResourceCPU:    "50000m", // 50 cpu
		corev1.ResourceMemory: "500Gi",  // 500Gi
	}
)

func TestDryRunPreemption(test *testing.T) {
	fixtures := []struct {
		name                    string
		registerPlugins         []schedulertesting.RegisterPluginFunc
		nodeNames               []string
		testPods                []*corev1.Pod
		initPods                []*corev1.Pod
		expected                [][]Candidate
		expectedNumFilterCalled []int32
		fakeFilterRC            framework.Code // return code for fake filter plugin
	}{
		{
			name: "a pod that does not fit on any node",
			registerPlugins: []schedulertesting.RegisterPluginFunc{
				schedulertesting.RegisterFilterPlugin("FalseFilter", schedulertesting.NewFalseFilterPlugin),
			},
			nodeNames: []string{"node1", "node2"},
			testPods: []*corev1.Pod{
				schedulertesting.MakePod().Name("p").UID("p").Priority(highPriority).Obj(),
			},
			initPods: []*corev1.Pod{
				schedulertesting.MakePod().Name("p1").UID("p1").Node("node1").Priority(midPriority).Obj(),
				schedulertesting.MakePod().Name("p2").UID("p2").Node("node2").Priority(midPriority).Obj(),
			},
			expected:                [][]Candidate{{}},
			expectedNumFilterCalled: []int32{2},
		},
	}

	labelKeys := []string{"hostname", "zone", "region"}
	for _, fixture := range fixtures {
		test.Run(fixture.name, func(t *testing.T) {
			// prepare node-pod snapshot for Framework
			nodes := make([]*corev1.Node, len(fixture.nodeNames))
			fakeFilterRCMap := make(map[string]framework.Code, len(fixture.nodeNames))
			for i, nodeName := range fixture.nodeNames {
				nodeWrapper := schedulertesting.MakeNode().Capacity(veryLargeRes)
				tpKeys := strings.Split(nodeName, "/")
				nodeWrapper.Name(tpKeys[0])
				for i, labelVal := range strings.Split(nodeName, "/") {
					nodeWrapper.Label(labelKeys[i], labelVal)
				}
				nodes[i] = nodeWrapper.Obj()
				fakeFilterRCMap[nodeName] = fixture.fakeFilterRC
			}
			snapshot := internalcache.NewSnapshot(fixture.initPods, nodes)

			fwk, err := schedulertesting.NewFramework()

			preemption := &DefaultPreemption{
				framework: fwk,
				args:      configv1.DefaultPreemptionArgs{},
				podLister: nil,
			}

			for cycle, pod := range fixture.testPods {
				candidates, _, _ := preemption.DryRunPreemption()
				// Sort the values (inner victims) and the candidate itself (by its NominatedNodeName).
				sort.Slice(candidates, func(i, j int) bool {
					return candidates[i].Name() < candidates[j].Name()
				})
				for i := range candidates {
					victims := candidates[i].Victims().Pods
					sort.Slice(victims, func(i, j int) bool {
						return victims[i].Name < victims[j].Name
					})
				}
				var candidates []Candidate
				for i := range candidates {
					candidates = append(candidates, Candidate{victims: candidates[i].Victims(), name: candidates[i].Name()})
				}

				if fakePlugin.NumFilterCalled-prevNumFilterCalled != fixture.expectedNumFilterCalled[cycle] {
					t.Errorf("cycle %d: got NumFilterCalled=%d, want %d", cycle,
						fakePlugin.NumFilterCalled-prevNumFilterCalled, fixture.expectedNumFilterCalled[cycle])
				}
				prevNumFilterCalled = fakePlugin.NumFilterCalled
				if diff := cmp.Diff(fixture.expected[cycle], candidates, cmp.AllowUnexported(Candidate{})); diff != "" {
					t.Errorf("cycle %d: unexpected candidates (-want, +got): %s", cycle, diff)
				}
			}
		})
	}

}
