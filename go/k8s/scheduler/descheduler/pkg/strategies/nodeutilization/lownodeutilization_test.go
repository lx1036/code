package nodeutilization

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"k8s-lx1036/k8s/scheduler/descheduler/pkg/api"
	"k8s-lx1036/k8s/scheduler/descheduler/pkg/evictions"
	"k8s-lx1036/k8s/scheduler/descheduler/pkg/test"

	v1 "k8s.io/api/core/v1"
	"k8s.io/api/policy/v1beta1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	core "k8s.io/client-go/testing"
	"k8s.io/kubernetes/pkg/apis/scheduling"
)

var (
	lowPriority  = int32(0)
	highPriority = int32(10000)
)

// INFO: 可以直接复制或模仿
func TestLowNodeUtilization(t *testing.T) {
	ctx := context.Background()
	n1NodeName := "n1"
	n2NodeName := "n2"
	n3NodeName := "n3"
	fixtures := []struct {
		name                         string
		thresholds, targetThresholds api.ResourceThresholds
		nodes                        map[string]*v1.Node
		pods                         map[string]*v1.PodList
		maxPodsToEvictPerNode        int
		expectedPodsEvicted          int
		evictedPods                  []string
	}{
		{
			name: "no evictable pods",
			thresholds: api.ResourceThresholds{
				v1.ResourceCPU:  30,
				v1.ResourcePods: 30,
			},
			targetThresholds: api.ResourceThresholds{
				v1.ResourceCPU:  50,
				v1.ResourcePods: 50,
			},
			nodes: map[string]*v1.Node{
				n1NodeName: test.BuildTestNode(n1NodeName, 4000, 3000, 9, nil),
				n2NodeName: test.BuildTestNode(n2NodeName, 4000, 3000, 10, nil),
				n3NodeName: test.BuildTestNode(n3NodeName, 4000, 3000, 10, test.SetNodeUnschedulable),
			},
			pods: map[string]*v1.PodList{
				n1NodeName: { // cpu_request_total 已经超过 4000 * 50%，属于高负载node
					Items: []v1.Pod{
						// INFO: 以下三种pods不驱逐: daemonset pod, pod with local storage, 以及高优先级pod
						// These won't be evicted.
						*test.BuildTestPod("p1", 400, 0, n1NodeName, test.SetDaemonsetOwnerRef),
						*test.BuildTestPod("p2", 400, 0, n1NodeName, test.SetDaemonsetOwnerRef),
						*test.BuildTestPod("p3", 400, 0, n1NodeName, test.SetDaemonsetOwnerRef),
						*test.BuildTestPod("p4", 400, 0, n1NodeName, test.SetDaemonsetOwnerRef),
						*test.BuildTestPod("p5", 400, 0, n1NodeName, func(pod *v1.Pod) {
							// A pod with local storage.
							test.SetNormalOwnerRef(pod)
							pod.Spec.Volumes = []v1.Volume{
								{
									Name: "sample",
									VolumeSource: v1.VolumeSource{
										HostPath: &v1.HostPathVolumeSource{Path: "somePath"},
										EmptyDir: &v1.EmptyDirVolumeSource{
											SizeLimit: resource.NewQuantity(int64(10), resource.BinarySI)},
									},
								},
							}
							// A Mirror Pod.
							pod.Annotations = test.GetMirrorPodAnnotation()
						}),
						*test.BuildTestPod("p6", 400, 0, n1NodeName, func(pod *v1.Pod) {
							// A Critical Pod.
							pod.Namespace = "kube-system"
							priority := scheduling.SystemCriticalPriority
							pod.Spec.Priority = &priority
						}),
					},
				},
				n2NodeName: {
					Items: []v1.Pod{
						*test.BuildTestPod("p9", 400, 0, n1NodeName, test.SetReplicaSetOwnerRef),
					},
				},
				n3NodeName: {},
			},
			maxPodsToEvictPerNode: 0,
			expectedPodsEvicted:   0,
		},
		{
			name: "without priorities",
			thresholds: api.ResourceThresholds{
				v1.ResourceCPU:  30,
				v1.ResourcePods: 30,
			},
			targetThresholds: api.ResourceThresholds{
				v1.ResourceCPU:  50,
				v1.ResourcePods: 50,
			},
			nodes: map[string]*v1.Node{
				n1NodeName: test.BuildTestNode(n1NodeName, 4000, 3000, 9, nil),
				n2NodeName: test.BuildTestNode(n2NodeName, 4000, 3000, 10, nil),
				n3NodeName: test.BuildTestNode(n3NodeName, 4000, 3000, 10, test.SetNodeUnschedulable),
			},
			pods: map[string]*v1.PodList{
				n1NodeName: { // cpu_request_total 已经超过 4000 * 50%，属于高负载node
					Items: []v1.Pod{ // INFO: 驱逐四个pod，这样cpu使用了 1600，低于 2000
						*test.BuildTestPod("p1", 400, 0, n1NodeName, test.SetReplicaSetOwnerRef),
						*test.BuildTestPod("p2", 400, 0, n1NodeName, test.SetReplicaSetOwnerRef),
						*test.BuildTestPod("p3", 400, 0, n1NodeName, test.SetReplicaSetOwnerRef),
						*test.BuildTestPod("p4", 400, 0, n1NodeName, test.SetReplicaSetOwnerRef),
						*test.BuildTestPod("p5", 400, 0, n1NodeName, test.SetReplicaSetOwnerRef),
						// These won't be evicted.
						*test.BuildTestPod("p6", 400, 0, n1NodeName, test.SetDaemonsetOwnerRef),
						*test.BuildTestPod("p7", 400, 0, n1NodeName, func(pod *v1.Pod) {
							// A pod with local storage.
							test.SetNormalOwnerRef(pod)
							pod.Spec.Volumes = []v1.Volume{
								{
									Name: "sample",
									VolumeSource: v1.VolumeSource{
										HostPath: &v1.HostPathVolumeSource{Path: "somePath"},
										EmptyDir: &v1.EmptyDirVolumeSource{
											SizeLimit: resource.NewQuantity(int64(10), resource.BinarySI)},
									},
								},
							}
							// A Mirror Pod.
							pod.Annotations = test.GetMirrorPodAnnotation()
						}),
						*test.BuildTestPod("p8", 400, 0, n1NodeName, func(pod *v1.Pod) {
							// A Critical Pod.
							pod.Namespace = "kube-system"
							priority := scheduling.SystemCriticalPriority
							pod.Spec.Priority = &priority
						}),
					},
				},
				n2NodeName: {
					Items: []v1.Pod{
						*test.BuildTestPod("p9", 400, 0, n1NodeName, test.SetReplicaSetOwnerRef),
					},
				},
				n3NodeName: {},
			},
			maxPodsToEvictPerNode: 0,
			expectedPodsEvicted:   4,
		},
		{
			name: "without priorities when cpu is exhausted",
			thresholds: api.ResourceThresholds{
				v1.ResourceCPU:  30,
				v1.ResourcePods: 30,
			},
			targetThresholds: api.ResourceThresholds{
				v1.ResourceCPU:  50,
				v1.ResourcePods: 50,
			},
			nodes: map[string]*v1.Node{
				n1NodeName: test.BuildTestNode(n1NodeName, 4000, 3000, 9, nil),
				n2NodeName: test.BuildTestNode(n2NodeName, 4000, 3000, 10, nil),
				n3NodeName: test.BuildTestNode(n3NodeName, 4000, 3000, 10, test.SetNodeUnschedulable),
			},
			pods: map[string]*v1.PodList{
				n1NodeName: { // cpu_request_total 已经超过 4000 * 50%，属于高负载node
					Items: []v1.Pod{ // INFO: 驱逐四个pod，这样cpu使用了 1600，低于 2000
						*test.BuildTestPod("p1", 400, 300, n1NodeName, test.SetReplicaSetOwnerRef),
						*test.BuildTestPod("p2", 400, 300, n1NodeName, test.SetReplicaSetOwnerRef),
						*test.BuildTestPod("p3", 400, 300, n1NodeName, test.SetReplicaSetOwnerRef),
						*test.BuildTestPod("p4", 400, 300, n1NodeName, test.SetReplicaSetOwnerRef),
						*test.BuildTestPod("p5", 400, 300, n1NodeName, test.SetReplicaSetOwnerRef),
						// These won't be evicted.
						*test.BuildTestPod("p6", 400, 300, n1NodeName, test.SetDaemonsetOwnerRef),
						*test.BuildTestPod("p7", 400, 300, n1NodeName, func(pod *v1.Pod) {
							// A pod with local storage.
							test.SetNormalOwnerRef(pod)
							pod.Spec.Volumes = []v1.Volume{
								{
									Name: "sample",
									VolumeSource: v1.VolumeSource{
										HostPath: &v1.HostPathVolumeSource{Path: "somePath"},
										EmptyDir: &v1.EmptyDirVolumeSource{
											SizeLimit: resource.NewQuantity(int64(10), resource.BinarySI)},
									},
								},
							}
							// A Mirror Pod.
							pod.Annotations = test.GetMirrorPodAnnotation()
						}),
						*test.BuildTestPod("p8", 400, 300, n1NodeName, func(pod *v1.Pod) {
							// A Critical Pod.
							pod.Namespace = "kube-system"
							priority := scheduling.SystemCriticalPriority
							pod.Spec.Priority = &priority
						}),
					},
				},
				n2NodeName: {
					Items: []v1.Pod{
						*test.BuildTestPod("p9", 400, 2100, n1NodeName, test.SetReplicaSetOwnerRef),
					},
				},
				n3NodeName: {},
			},
			maxPodsToEvictPerNode: 0,
			expectedPodsEvicted:   4,
		},
	}

	for _, fixture := range fixtures {
		t.Run(fixture.name, func(t *testing.T) {
			fakeClient := &fake.Clientset{}
			fakeClient.Fake.AddReactor("list", "pods", func(action core.Action) (bool, runtime.Object, error) {
				list := action.(core.ListAction)
				fieldString := list.GetListRestrictions().Fields.String()
				if strings.Contains(fieldString, n1NodeName) {
					return true, fixture.pods[n1NodeName], nil
				}
				if strings.Contains(fieldString, n2NodeName) {
					return true, fixture.pods[n2NodeName], nil
				}
				if strings.Contains(fieldString, n3NodeName) {
					return true, fixture.pods[n3NodeName], nil
				}
				return true, nil, fmt.Errorf("failed to list: %v", list)
			})

			fakeClient.Fake.AddReactor("get", "nodes", func(action core.Action) (bool, runtime.Object, error) {
				getAction := action.(core.GetAction)
				if node, exists := fixture.nodes[getAction.GetName()]; exists {
					return true, node, nil
				}
				return true, nil, fmt.Errorf("wrong node: %v", getAction.GetName())
			})

			podsForEviction := make(map[string]struct{})
			for _, pod := range fixture.evictedPods {
				podsForEviction[pod] = struct{}{}
			}
			evictionFailed := false
			if len(fixture.evictedPods) > 0 {
				fakeClient.Fake.AddReactor("create", "pods", func(action core.Action) (bool, runtime.Object, error) {
					getAction := action.(core.CreateAction)
					obj := getAction.GetObject()
					if eviction, ok := obj.(*v1beta1.Eviction); ok { // 这里为何是 v1beta1.Eviction 对象，不是 create pods 么?
						if _, exists := podsForEviction[eviction.Name]; exists {
							return true, obj, nil
						}
						evictionFailed = true
						return true, nil, fmt.Errorf("pod %q was unexpectedly evicted", eviction.Name)
					}
					return true, obj, nil
				})
			}

			var nodes []*v1.Node
			for _, node := range fixture.nodes {
				nodes = append(nodes, node)
			}
			podEvictor := evictions.NewPodEvictor(
				fakeClient,
				"v1",
				false,
				fixture.maxPodsToEvictPerNode,
				nodes,
				false)
			strategy := api.DeschedulerStrategy{
				Enabled: true,
				//Weight:  0,
				Params: &api.StrategyParameters{
					NodeResourceUtilizationThresholds: &api.NodeResourceUtilizationThresholds{
						Thresholds:       fixture.thresholds,
						TargetThresholds: fixture.targetThresholds,
					},
				},
			}

			// 执行业务逻辑
			LowNodeUtilization(ctx, fakeClient, strategy, nodes, podEvictor)

			podsEvicted := podEvictor.TotalEvicted()
			if fixture.expectedPodsEvicted != podsEvicted {
				t.Errorf("Expected %#v pods to be evicted but %#v got evicted", fixture.expectedPodsEvicted, podsEvicted)
			}
			if evictionFailed {
				t.Errorf("Pod evictions failed unexpectedly")
			}
		})
	}
}
