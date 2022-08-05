package defaultpreemption

import (
	"context"
	"k8s.io/client-go/tools/events"
	"sort"
	"strings"
	"testing"

	configv1 "k8s-lx1036/k8s/scheduler/pkg/apis/config/v1"
	"k8s-lx1036/k8s/scheduler/pkg/framework"
	"k8s-lx1036/k8s/scheduler/pkg/framework/parallelize"
	"k8s-lx1036/k8s/scheduler/pkg/framework/plugins/defaultbinder"
	"k8s-lx1036/k8s/scheduler/pkg/framework/plugins/noderesources"
	"k8s-lx1036/k8s/scheduler/pkg/framework/plugins/queuesort"
	frameworkruntime "k8s-lx1036/k8s/scheduler/pkg/framework/runtime"
	internalcache "k8s-lx1036/k8s/scheduler/pkg/internal/cache"
	internalqueue "k8s-lx1036/k8s/scheduler/pkg/internal/queue"
	schedulertesting "k8s-lx1036/k8s/scheduler/pkg/testing"

	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/informers"
	clientsetfake "k8s.io/client-go/kubernetes/fake"
)

var (
	negPriority, lowPriority, midPriority, highPriority, veryHighPriority = int32(-100), int32(0), int32(100), int32(1000), int32(10000)

	veryLargeRes = map[corev1.ResourceName]string{
		corev1.ResourceCPU:    "50000m", // 50 cpu
		corev1.ResourceMemory: "500Gi",  // 500Gi
	}
)

func getDefaultDefaultPreemptionArgs() *configv1.DefaultPreemptionArgs {
	return &configv1.DefaultPreemptionArgs{
		MinCandidateNodesPercentage: 10,
		MinCandidateNodesAbsolute:   100,
	}
}

func TestDryRunPreemption(test *testing.T) {
	fixtures := []struct {
		name                    string
		registerPlugins         []schedulertesting.RegisterPluginFunc
		nodeNames               []string
		testPods                []*corev1.Pod // 开始抢占 initPods
		initPods                []*corev1.Pod
		expected                [][]Candidate
		expectedNumFilterCalled []int32
		fakeFilterRC            framework.Code // return code for fake filter plugin
		disableParallelism      bool
		args                    *configv1.DefaultPreemptionArgs
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
			// INFO: new framework，可以参考
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
			snapshot := internalcache.NewSnapshot(fixture.initPods, nodes) // 集群 pod-node 初始状态
			fakePlugin := schedulertesting.FakeFilterPlugin{
				FailedNodeReturnCodeMap: fakeFilterRCMap,
			}
			registeredPlugins := []schedulertesting.RegisterPluginFunc{
				schedulertesting.RegisterFilterPlugin("FakePlugin", func(configuration runtime.Object, f *frameworkruntime.Framework) (framework.Plugin, error) {
					return &fakePlugin, nil
				}),
				schedulertesting.RegisterQueueSortPlugin(queuesort.Name, queuesort.New),
				schedulertesting.RegisterBindPlugin(defaultbinder.Name, defaultbinder.New),
			}
			registeredPlugins = append(registeredPlugins, fixture.registerPlugins...)
			parallelism := parallelize.DefaultParallelism
			if fixture.disableParallelism {
				parallelism = 1
			}
			var objs []runtime.Object
			for _, p := range append(fixture.testPods, fixture.initPods...) {
				objs = append(objs, p)
			}
			for _, n := range nodes {
				objs = append(objs, n)
			}
			informerFactory := informers.NewSharedInformerFactory(clientsetfake.NewSimpleClientset(objs...), 0) // INFO: 这里可以重点参考 informerFactory
			fwk, err := schedulertesting.NewFramework(registeredPlugins, "",
				frameworkruntime.WithSnapshotSharedLister(snapshot),
				frameworkruntime.WithInformerFactory(informerFactory),
				frameworkruntime.WithParallelism(parallelism),
				frameworkruntime.WithPodNominator(internalqueue.NewPodNominator(informerFactory.Core().V1().Pods().Lister())),
			)
			if err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			informerFactory.Start(ctx.Done())
			informerFactory.WaitForCacheSync(ctx.Done()) // 这里的 informerFactory 是 fake 的

			// new Preemption
			nodeInfos, err := snapshot.NodeInfos().List() // INFO: 全局总共所有的 nodes
			if err != nil {
				t.Fatal(err)
			}
			sort.Slice(nodeInfos, func(i, j int) bool {
				return nodeInfos[i].Node().Name < nodeInfos[j].Node().Name
			})
			if fixture.args == nil {
				fixture.args = getDefaultDefaultPreemptionArgs()
			}
			preemption := &DefaultPreemption{
				framework: fwk,
				args:      *fixture.args,
				podLister: informerFactory.Core().V1().Pods().Lister(),
			}

			// do preemption，一个个 pod 开始抢占
			var prevNumFilterCalled int32
			for cycle, pod := range fixture.testPods {
				cycleState := framework.NewCycleState()
				// Some tests rely on PreFilter plugin to compute its CycleState.
				if _, status := fwk.RunPreFilterPlugins(context.Background(), cycleState, pod); !status.IsSuccess() {
					t.Errorf("cycle %d: Unexpected PreFilter Status: %v", cycle, status)
				}
				offset, numCandidates := preemption.GetOffsetAndNumCandidates(int32(len(nodeInfos))) // 这里假设 numCandidates=100台，offset=10台
				// INFO: !!!这里直接测试抢占逻辑!!!
				candidates, _, _ := preemption.DryRunPreemption(ctx, pod, nodeInfos, offset, numCandidates, cycleState)
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
				var candidatesCopy []Candidate
				for i := range candidates {
					candidatesCopy = append(candidatesCopy, Candidate{victims: candidates[i].Victims(), name: candidates[i].Name()})
				}

				// check
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

func TestPostFilter(test *testing.T) {

	for _, fixture := range fixtures {
		test.Run(fixture.name, func(t *testing.T) {
			cs := clientsetfake.NewSimpleClientset()

			registeredPlugins := []schedulertesting.RegisterPluginFunc{
				schedulertesting.RegisterPluginAsExtensions(noderesources.Name, noderesources.NewFit, "Filter", "PreFilter"),
				schedulertesting.RegisterPluginAsExtensions("test-plugin", schedulertesting.NewTestPlugin, "PreFilter"),
				schedulertesting.RegisterQueueSortPlugin(queuesort.Name, queuesort.New),
				schedulertesting.RegisterBindPlugin(defaultbinder.Name, defaultbinder.New),
			}
			informerFactory := informers.NewSharedInformerFactory(clientsetfake.NewSimpleClientset(), 0) // INFO: 注意这里是空的，和上面的不一样
			fwk, err := schedulertesting.NewFramework(registeredPlugins, "",
				frameworkruntime.WithClientSet(cs),
				frameworkruntime.WithEventRecorder(&events.FakeRecorder{}),
				frameworkruntime.WithSnapshotSharedLister(snapshot),
				frameworkruntime.WithInformerFactory(informerFactory),
				frameworkruntime.WithParallelism(parallelize.DefaultParallelism),
				frameworkruntime.WithPodNominator(internalqueue.NewPodNominator(informerFactory.Core().V1().Pods().Lister())),
			)
			if err != nil {
				t.Fatal(err)
			}

			preemption := &DefaultPreemption{
				framework: fwk,
				args:      *getDefaultDefaultPreemptionArgs(),
				podLister: informerFactory.Core().V1().Pods().Lister(),
			}

		})
	}
}
