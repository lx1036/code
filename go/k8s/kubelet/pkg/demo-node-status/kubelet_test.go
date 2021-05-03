package demo_node_status

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/clock"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/flowcontrol"
	cadvisortest "k8s.io/kubernetes/pkg/kubelet/cadvisor/testing"
	"k8s.io/kubernetes/pkg/kubelet/cm"
	"k8s.io/kubernetes/pkg/kubelet/config"
	"k8s.io/kubernetes/pkg/kubelet/configmap"
	kubecontainer "k8s.io/kubernetes/pkg/kubelet/container"
	containertest "k8s.io/kubernetes/pkg/kubelet/container/testing"
	"k8s.io/kubernetes/pkg/kubelet/logs"
	"k8s.io/kubernetes/pkg/kubelet/pleg"
	"k8s.io/kubernetes/pkg/kubelet/pluginmanager"
	kubepod "k8s.io/kubernetes/pkg/kubelet/pod"
	podtest "k8s.io/kubernetes/pkg/kubelet/pod/testing"
	proberesults "k8s.io/kubernetes/pkg/kubelet/prober/results"
	probetest "k8s.io/kubernetes/pkg/kubelet/prober/testing"
	"k8s.io/kubernetes/pkg/kubelet/secret"
	"k8s.io/kubernetes/pkg/kubelet/stats"
	"k8s.io/kubernetes/pkg/kubelet/status"
	statustest "k8s.io/kubernetes/pkg/kubelet/status/testing"
	"k8s.io/kubernetes/pkg/kubelet/util/queue"
	kubeletvolume "k8s.io/kubernetes/pkg/kubelet/volumemanager"
	"k8s.io/kubernetes/pkg/volume"
	volumetest "k8s.io/kubernetes/pkg/volume/testing"
	"k8s.io/kubernetes/pkg/volume/util/hostutil"
	"k8s.io/kubernetes/pkg/volume/util/subpath"
	"k8s.io/utils/mount"
)

const (
	testKubeletHostname = "127.0.0.1"
	testKubeletHostIP   = "127.0.0.1"
)

type TestKubelet struct {
	kubelet          *Kubelet
	fakeRuntime      *containertest.FakeRuntime
	fakeKubeClient   *fake.Clientset
	fakeMirrorClient *podtest.FakeMirrorClient
	fakeClock        *clock.FakeClock
	mounter          mount.Interface
	volumePlugin     *volumetest.FakeVolumePlugin
}

func (tk *TestKubelet) Cleanup() {
	if tk.kubelet != nil {
		os.RemoveAll(tk.kubelet.rootDirectory)
	}
}

// newTestKubelet returns test kubelet with two images.
func newTestKubelet(t *testing.T, controllerAttachDetachEnabled bool) *TestKubelet {
	imageList := []kubecontainer.Image{
		{
			ID:       "abc",
			RepoTags: []string{"k8s.gcr.io:v1", "k8s.gcr.io:v2"},
			Size:     123,
		},
		{
			ID:       "efg",
			RepoTags: []string{"k8s.gcr.io:v3", "k8s.gcr.io:v4"},
			Size:     456,
		},
	}

	return newTestKubeletWithImageList(t, imageList, controllerAttachDetachEnabled, true /*initFakeVolumePlugin*/)
}

func newTestKubeletWithImageList(t *testing.T, imageList []kubecontainer.Image, controllerAttachDetachEnabled bool, initFakeVolumePlugin bool) *TestKubelet {
	fakeRuntime := &containertest.FakeRuntime{}
	fakeRuntime.RuntimeType = "test"
	fakeRuntime.VersionInfo = "1.5.0"
	fakeRuntime.ImageList = imageList
	fakeRuntime.RuntimeStatus = &kubecontainer.RuntimeStatus{
		Conditions: []kubecontainer.RuntimeCondition{
			{Type: "RuntimeReady", Status: true},
			{Type: "NetworkReady", Status: true},
		},
	}

	fakeRecorder := &record.FakeRecorder{}
	fakeKubeClient := &fake.Clientset{}
	kubelet := &Kubelet{}
	kubelet.recorder = fakeRecorder
	kubelet.kubeClient = fakeKubeClient
	kubelet.heartbeatClient = fakeKubeClient
	kubelet.os = &containertest.FakeOS{}
	kubelet.mounter = mount.NewFakeMounter(nil)
	kubelet.hostutil = hostutil.NewFakeHostUtil(nil)
	kubelet.subpather = &subpath.FakeSubpath{}

	kubelet.hostname = testKubeletHostname
	kubelet.nodeName = types.NodeName(testKubeletHostname)
	kubelet.runtimeState = newRuntimeState(maxWaitForContainerRuntime)
	kubelet.runtimeState.setNetworkState(nil)
	if tempDir, err := ioutil.TempDir("", "kubelet_test."); err != nil {
		t.Fatalf("can't make a temp rootdir: %v", err)
	} else {
		kubelet.rootDirectory = tempDir
	}
	if err := os.MkdirAll(kubelet.rootDirectory, 0750); err != nil {
		t.Fatalf("can't mkdir(%q): %v", kubelet.rootDirectory, err)
	}
	kubelet.sourcesReady = config.NewSourcesReady(func(_ sets.String) bool { return true })
	kubelet.masterServiceNamespace = metav1.NamespaceDefault
	//kubelet.serviceLister = testServiceLister{}
	kubelet.serviceHasSynced = func() bool { return true }
	kubelet.nodeHasSynced = func() bool { return true }
	kubelet.nodeLister = testNodeLister{
		nodes: []*v1.Node{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: string(kubelet.nodeName),
				},
				Status: v1.NodeStatus{
					Conditions: []v1.NodeCondition{
						{
							Type:    v1.NodeReady,
							Status:  v1.ConditionTrue,
							Reason:  "Ready",
							Message: "Node ready",
						},
					},
					Addresses: []v1.NodeAddress{
						{
							Type:    v1.NodeInternalIP,
							Address: testKubeletHostIP,
						},
					},
				},
			},
		},
	}
	kubelet.recorder = fakeRecorder
	if err := kubelet.setupDataDirs(); err != nil {
		t.Fatalf("can't initialize kubelet data dirs: %v", err)
	}
	kubelet.daemonEndpoints = &v1.NodeDaemonEndpoints{}

	kubelet.cadvisor = &cadvisortest.Fake{}
	machineInfo, _ := kubelet.cadvisor.MachineInfo()
	kubelet.setCachedMachineInfo(machineInfo)

	fakeMirrorClient := podtest.NewFakeMirrorClient()
	secretManager := secret.NewSimpleSecretManager(kubelet.kubeClient)
	kubelet.secretManager = secretManager
	configMapManager := configmap.NewSimpleConfigMapManager(kubelet.kubeClient)
	kubelet.configMapManager = configMapManager
	kubelet.podManager = kubepod.NewBasicPodManager(fakeMirrorClient, kubelet.secretManager, kubelet.configMapManager)
	kubelet.statusManager = status.NewManager(fakeKubeClient, kubelet.podManager, &statustest.FakePodDeletionSafetyProvider{})

	kubelet.containerRuntime = fakeRuntime
	kubelet.runtimeCache = containertest.NewFakeRuntimeCache(kubelet.containerRuntime)
	//kubelet.reasonCache = NewReasonCache()
	kubelet.podCache = containertest.NewFakeCache(kubelet.containerRuntime)
	/*kubelet.podWorkers = &fakePodWorkers{
		syncPodFn: kubelet.syncPod,
		cache:     kubelet.podCache,
		t:         t,
	}*/

	kubelet.probeManager = probetest.FakeManager{}
	kubelet.livenessManager = proberesults.NewManager()
	kubelet.startupManager = proberesults.NewManager()

	kubelet.containerManager = cm.NewStubContainerManager()
	/*fakeNodeRef := &v1.ObjectReference{
		Kind:      "Node",
		Name:      testKubeletHostname,
		UID:       types.UID(testKubeletHostname),
		Namespace: "",
	}*/

	//volumeStatsAggPeriod := time.Second * 10
	//kubelet.resourceAnalyzer = serverstats.NewResourceAnalyzer(kubelet, volumeStatsAggPeriod)

	kubelet.StatsProvider = stats.NewCadvisorStatsProvider(
		kubelet.cadvisor,
		kubelet.resourceAnalyzer,
		kubelet.podManager,
		kubelet.runtimeCache,
		fakeRuntime,
		kubelet.statusManager)
	/*fakeImageGCPolicy := images.ImageGCPolicy{
		HighThresholdPercent: 90,
		LowThresholdPercent:  80,
	}*/
	//imageGCManager, err := images.NewImageGCManager(fakeRuntime, kubelet.StatsProvider, fakeRecorder, fakeNodeRef, fakeImageGCPolicy, "")
	//assert.NoError(t, err)
	//kubelet.imageManager = &fakeImageGCManager{
	//	fakeImageService: fakeRuntime,
	//	ImageGCManager:   imageGCManager,
	//}
	kubelet.containerLogManager = logs.NewStubContainerLogManager()
	containerGCPolicy := kubecontainer.GCPolicy{
		MinAge:             time.Duration(0),
		MaxPerPodContainer: 1,
		MaxContainers:      -1,
	}
	containerGC, err := kubecontainer.NewContainerGC(fakeRuntime, containerGCPolicy, kubelet.sourcesReady)
	assert.NoError(t, err)
	kubelet.containerGC = containerGC

	fakeClock := clock.NewFakeClock(time.Now())
	kubelet.backOff = flowcontrol.NewBackOff(time.Second, time.Minute)
	kubelet.backOff.Clock = fakeClock
	//kubelet.podKiller = NewPodKiller(kubelet)
	kubelet.resyncInterval = 10 * time.Second
	kubelet.workQueue = queue.NewBasicWorkQueue(fakeClock)
	// Relist period does not affect the tests.
	kubelet.pleg = pleg.NewGenericPLEG(fakeRuntime, 100, time.Hour, nil, clock.RealClock{})
	kubelet.clock = fakeClock

	/*nodeRef := &v1.ObjectReference{
		Kind:      "Node",
		Name:      string(kubelet.nodeName),
		UID:       types.UID(kubelet.nodeName),
		Namespace: "",
	}
	etcHostsPathFunc := func(podUID types.UID) string {
		return ""
		//return getEtcHostsPath(kubelet.getPodDir(podUID))
	}*/
	// setup eviction manager
	/*evictionManager, evictionAdmitHandler := eviction.NewManager(kubelet.resourceAnalyzer, eviction.Config{},
	killPodNow(kubelet.podWorkers, fakeRecorder), kubelet.podManager.GetMirrorPodByPod,
	kubelet.imageManager, kubelet.containerGC, fakeRecorder, nodeRef, kubelet.clock, etcHostsPathFunc)
	kubelet.evictionManager = evictionManager
	kubelet.admitHandlers.AddPodAdmitHandler(evictionAdmitHandler)
	// Add this as cleanup predicate pod admitter
	kubelet.admitHandlers.AddPodAdmitHandler(lifecycle.NewPredicateAdmitHandler(kubelet.getNodeAnyWay,
		lifecycle.NewAdmissionFailureHandlerStub(), kubelet.containerManager.UpdatePluginResources))*/

	allPlugins := []volume.VolumePlugin{}
	plug := &volumetest.FakeVolumePlugin{PluginName: "fake", Host: nil}
	if initFakeVolumePlugin {
		allPlugins = append(allPlugins, plug)
	} else {
	}

	//var prober volume.DynamicPluginProber // TODO (#51147) inject mock
	/*kubelet.volumePluginMgr, err = NewInitializedVolumePluginMgr(kubelet, kubelet.secretManager, kubelet.configMapManager, token.NewManager(kubelet.kubeClient), allPlugins, prober)
	require.NoError(t, err, "Failed to initialize VolumePluginMgr")*/
	kubelet.volumeManager = kubeletvolume.NewVolumeManager(
		controllerAttachDetachEnabled,
		kubelet.nodeName,
		kubelet.podManager,
		kubelet.statusManager,
		fakeKubeClient,
		kubelet.volumePluginMgr,
		fakeRuntime,
		kubelet.mounter,
		kubelet.hostutil,
		kubelet.getPodsDir(),
		kubelet.recorder,
		false, /* experimentalCheckNodeCapabilitiesBeforeMount*/
		false, /* keepTerminatedPodVolumes */
		volumetest.NewBlockVolumePathHandler())
	kubelet.pluginManager = pluginmanager.NewPluginManager(
		kubelet.getPluginsRegistrationDir(), /* sockDir */
		kubelet.recorder,
	)
	//kubelet.setNodeStatusFuncs = kubelet.defaultNodeStatusFuncs()

	// enable active deadline handler
	/*activeDeadlineHandler, err := newActiveDeadlineHandler(kubelet.statusManager, kubelet.recorder, kubelet.clock)
	require.NoError(t, err, "Can't initialize active deadline handler")

	kubelet.AddPodSyncLoopHandler(activeDeadlineHandler)
	kubelet.AddPodSyncHandler(activeDeadlineHandler)*/

	return &TestKubelet{
		kubelet:          kubelet,
		fakeRuntime:      fakeRuntime,
		fakeKubeClient:   fakeKubeClient,
		fakeMirrorClient: fakeMirrorClient,
		fakeClock:        fakeClock,
		mounter:          nil,
		volumePlugin:     plug,
	}
}

type testNodeLister struct {
	nodes []*v1.Node
}

func (nl testNodeLister) Get(name string) (*v1.Node, error) {
	for _, node := range nl.nodes {
		if node.Name == name {
			return node, nil
		}
	}
	return nil, fmt.Errorf("node with name: %s does not exist", name)
}

func (nl testNodeLister) List(_ labels.Selector) (ret []*v1.Node, err error) {
	return nl.nodes, nil
}
