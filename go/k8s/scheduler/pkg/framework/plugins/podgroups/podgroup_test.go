package podgroups

import (
	"context"
	"k8s.io/klog/v2"
	"testing"
	"time"

	podgroupv1 "k8s-lx1036/k8s/scheduler/pkg/apis/podgroup/v1"
	pgclientsetfake "k8s-lx1036/k8s/scheduler/pkg/client/clientset/versioned/fake"
	"k8s-lx1036/k8s/scheduler/pkg/client/informers/externalversions"
	"k8s-lx1036/k8s/scheduler/pkg/framework"
	"k8s-lx1036/k8s/scheduler/pkg/framework/plugins/defaultbinder"
	podgrouptesting "k8s-lx1036/k8s/scheduler/pkg/framework/plugins/podgroups/testing"
	"k8s-lx1036/k8s/scheduler/pkg/framework/plugins/queuesort"
	frameworkruntime "k8s-lx1036/k8s/scheduler/pkg/framework/runtime"
	internalcache "k8s-lx1036/k8s/scheduler/pkg/internal/cache"
	schedulertesting "k8s-lx1036/k8s/scheduler/pkg/testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	clientsetfake "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/events"
)

func TestPermit(test *testing.T) {
	fixtures := []struct {
		name     string
		pod      *corev1.Pod
		expected framework.Code
	}{
		{
			name:     "pods do not belong to any podGroup",
			pod:      schedulertesting.MakePod().Name("p1").UID("p1").Obj(),
			expected: framework.Success,
		},
		{
			name:     "pods belong to a podGroup, Wait",
			pod:      schedulertesting.MakePod().Name("p1").Namespace("ns1").UID("p1").Label(podgroupv1.PodGroupLabel, "pg1").Obj(),
			expected: framework.Wait,
		},
		{
			name:     "pods belong to a podGroup, Allow",
			pod:      schedulertesting.MakePod().Name("p1").Namespace("ns1").UID("p1").Label(podgroupv1.PodGroupLabel, "pg2").Obj(),
			expected: framework.Success,
		},
	}

	ctx := context.Background()
	cs := pgclientsetfake.NewSimpleClientset()
	pgInformerFactory := externalversions.NewSharedInformerFactory(cs, 0)
	pgInformer := pgInformerFactory.PodGroup().V1().PodGroups()
	pgInformerFactory.Start(ctx.Done())

	pg1 := podgrouptesting.MakePodGroup("pg1", "ns1", 2, nil, nil)
	pg2 := podgrouptesting.MakePodGroup("pg2", "ns1", 1, nil, nil)
	pgInformer.Informer().GetStore().Add(pg1) // 这里直接往 cache 里填数据，不用 pgInformerFactory.WaitForCache 去 fake sync 了
	pgInformer.Informer().GetStore().Add(pg2)

	fakeClient := clientsetfake.NewSimpleClientset()
	informerFactory := informers.NewSharedInformerFactory(fakeClient, 0) // INFO: 这里可以重点参考 informerFactory
	podInformer := informerFactory.Core().V1().Pods()
	informerFactory.Start(ctx.Done())

	existingPods, allNodes := podgrouptesting.MakeNodesAndPods(map[string]string{"app": "nginx"}, 60, 30)
	snapshot := internalcache.NewSnapshot(existingPods, allNodes) // 集群 pod-node 初始状态，集群快照，可以加速
	registeredPlugins := []schedulertesting.RegisterPluginFunc{
		schedulertesting.RegisterQueueSortPlugin(queuesort.Name, queuesort.New),
		schedulertesting.RegisterBindPlugin(defaultbinder.Name, defaultbinder.New),
	}
	fwk, err := schedulertesting.NewFramework(registeredPlugins, "",
		frameworkruntime.WithClientSet(fakeClient),
		frameworkruntime.WithSnapshotSharedLister(snapshot),
		frameworkruntime.WithInformerFactory(informerFactory),
		frameworkruntime.WithEventRecorder(&events.FakeRecorder{}),
	)
	if err != nil {
		test.Fatal(err)
	}

	scheduleDuration := 10 * time.Second
	for _, fixture := range fixtures {
		test.Run(fixture.name, func(t *testing.T) {
			podGroupManager := NewPodGroupManager(cs, snapshot, scheduleDuration, pgInformer, podInformer)
			pl := &PodGroup{
				framework:       fwk,
				podGroupManager: podGroupManager,
				scheduleTimeout: scheduleDuration,
			}
			status, _ := pl.Permit(context.Background(), framework.NewCycleState(), fixture.pod, "node1")
			if status.Code() != fixture.expected {
				t.Errorf("expected %v, got %v", fixture.expected, status.Code())
			}
		})
	}
}

func TestName(test *testing.T) {
	ctx := context.Background()
	fakeClient := clientsetfake.NewSimpleClientset()
	informerFactory := informers.NewSharedInformerFactory(fakeClient, 0) // INFO: 这里可以重点参考 informerFactory
	podInformer := informerFactory.Core().V1().Pods()
	podInformer.Lister() // podInformer.Lister() 加了这个 WaitForCacheSync 就不会阻塞了，表示已经做了初始 list
	informerFactory.Start(ctx.Done())
	if !cache.WaitForCacheSync(ctx.Done(), podInformer.Informer().HasSynced) {
		test.Fatal("WaitForCacheSync failed")
	}

	klog.Info("success")
}

func TestPostFilter(test *testing.T) {
	ctx := context.Background()
	cs := pgclientsetfake.NewSimpleClientset()
	pgInformerFactory := externalversions.NewSharedInformerFactory(cs, 0)
	pgInformer := pgInformerFactory.PodGroup().V1().PodGroups()
	pgInformerFactory.Start(ctx.Done())
	//pgInformerFactory.WaitForCacheSync(ctx.Done()) // pgInformer.Lister() 加了这个 WaitForCacheSync 就不会阻塞了
	pg1 := podgrouptesting.MakePodGroup("pg", "ns1", 2, nil, nil)
	pgInformer.Informer().GetStore().Add(pg1) // 这里直接往 cache 里填数据，不用 pgInformerFactory.WaitForCache 去 fake sync 了

	fakeClient := clientsetfake.NewSimpleClientset()
	informerFactory := informers.NewSharedInformerFactory(fakeClient, 0) // INFO: 这里可以重点参考 informerFactory
	podInformer := informerFactory.Core().V1().Pods()
	informerFactory.Start(ctx.Done())
	existingPods, allNodes := podgrouptesting.MakeNodesAndPods(map[string]string{"app": "nginx"}, 60, 30)
	snapshot := internalcache.NewSnapshot(existingPods, allNodes) // 集群 pod-node 初始状态，集群快照，可以加速
	registeredPlugins := []schedulertesting.RegisterPluginFunc{
		schedulertesting.RegisterQueueSortPlugin(queuesort.Name, queuesort.New),
		schedulertesting.RegisterBindPlugin(defaultbinder.Name, defaultbinder.New),
	}
	fwk, err := schedulertesting.NewFramework(registeredPlugins, "",
		frameworkruntime.WithClientSet(fakeClient),
		frameworkruntime.WithSnapshotSharedLister(snapshot),
		frameworkruntime.WithInformerFactory(informerFactory),
		frameworkruntime.WithEventRecorder(&events.FakeRecorder{}),
	)
	if err != nil {
		test.Fatal(err)
	}

	existingPods, allNodes = podgrouptesting.MakeNodesAndPods(map[string]string{podgroupv1.PodGroupLabel: "pg"}, 10, 30)
	for _, pod := range existingPods {
		pod.Namespace = "ns1"
	}
	groupPodSnapshot := internalcache.NewSnapshot(existingPods, allNodes) // 集群 pod-node 初始状态，集群快照，可以加速
	fixtures := []struct {
		name                 string
		pod                  *corev1.Pod
		expectedEmptyMsg     bool
		snapshotSharedLister framework.SharedLister
	}{
		{
			name:             "pod does not belong to any pod group",
			pod:              schedulertesting.MakePod().Name("p1").Namespace("ns1").UID("p1").Obj(),
			expectedEmptyMsg: false,
		},
		{
			name:                 "enough pods assigned, do not reject all", // 根据 groupPodSnapshot 30 台 nodes 已有 pg pod-group 的 10 个 pod 了，足够了
			pod:                  schedulertesting.MakePod().Name("p1").Namespace("ns1").UID("p1").Label(podgroupv1.PodGroupLabel, "pg").Obj(),
			expectedEmptyMsg:     true,
			snapshotSharedLister: groupPodSnapshot,
		},
		{
			name:             "pod failed at filter phase, reject all pods", // 根据 snapshot，30 台 nodes 60 个 pods，但是没有 pg pod-group 的 pod，加上 p1，数量不足
			pod:              schedulertesting.MakePod().Name("p1").Namespace("ns1").UID("p1").Label(podgroupv1.PodGroupLabel, "pg").Obj(),
			expectedEmptyMsg: false,
		},
	}

	scheduleDuration := 10 * time.Second
	nodeStatusMap := framework.NodeToStatusMap{"node1": framework.NewStatus(framework.Success, "")}
	for _, fixture := range fixtures {
		test.Run(fixture.name, func(t *testing.T) {
			var mgrSnapShot framework.SharedLister
			mgrSnapShot = snapshot
			if fixture.snapshotSharedLister != nil {
				mgrSnapShot = fixture.snapshotSharedLister
			}
			podGroupManager := NewPodGroupManager(cs, mgrSnapShot, scheduleDuration, pgInformer, podInformer)
			pl := &PodGroup{
				framework:       fwk,
				podGroupManager: podGroupManager,
				scheduleTimeout: scheduleDuration,
			}
			_, status := pl.PostFilter(context.Background(), framework.NewCycleState(), fixture.pod, nodeStatusMap)
			if (status.Message() == "") != fixture.expectedEmptyMsg {
				t.Errorf("expected %v, got %v", fixture.expectedEmptyMsg, status.Message() == "")
			}
		})
	}
}
