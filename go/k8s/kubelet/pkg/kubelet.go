package pkg

import (
	"fmt"
	"time"

	"k8s-lx1036/k8s/kubelet/pkg/cadvisor"
	"k8s-lx1036/k8s/kubelet/pkg/eviction"
	"k8s-lx1036/k8s/kubelet/pkg/images"
	serverstats "k8s-lx1036/k8s/kubelet/pkg/server/stats"
	"k8s-lx1036/k8s/kubelet/pkg/stats"
	"k8s-lx1036/k8s/kubelet/pkg/util/queue"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/clock"
	"k8s.io/client-go/informers"
	clientset "k8s.io/client-go/kubernetes"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/record"
	api "k8s.io/kubernetes/pkg/apis/core"
	"k8s.io/kubernetes/pkg/kubelet/configmap"
	kubecontainer "k8s.io/kubernetes/pkg/kubelet/container"
	kubepod "k8s.io/kubernetes/pkg/kubelet/pod"
	"k8s.io/kubernetes/pkg/kubelet/secret"
)

const (
	// backOffPeriod is the period to back off when pod syncing results in an
	// error. It is also used as the base period for the exponential backoff
	// container restarts and image pulls.
	backOffPeriod = time.Second * 10
)

type KubeletConfiguration struct {
	// This flag specifies the various Node Allocatable enforcements that Kubelet needs to perform.
	// This flag accepts a list of options. Acceptable options are `pods`, `system-reserved` & `kube-reserved`.
	// Refer to [Node Allocatable](https://git.k8s.io/community/contributors/design-proposals/node/node-allocatable.md) doc for more information.
	EnforceNodeAllocatable []string

	// Map of signal names to quantities that defines hard eviction thresholds. For example: {"memory.available": "300Mi"}.
	EvictionHard map[string]string
	// Map of signal names to quantities that defines soft eviction thresholds.  For example: {"memory.available": "300Mi"}.
	EvictionSoft map[string]string
	// Map of signal names to quantities that defines grace periods for each soft eviction signal. For example: {"memory.available": "30s"}.
	EvictionSoftGracePeriod map[string]string
	// Duration for which the kubelet has to wait before transitioning out of an eviction pressure condition.
	EvictionPressureTransitionPeriod metav1.Duration
	// Maximum allowed grace period (in seconds) to use when terminating pods in response to a soft eviction threshold being met.
	EvictionMaxPodGracePeriod int32
	// Map of signal names to quantities that defines minimum reclaims, which describe the minimum
	// amount of a given resource the kubelet will reclaim when performing a pod eviction while
	// that resource is under pressure. For example: {"imagefs.available": "2Gi"}
	EvictionMinimumReclaim map[string]string

	// How frequently to calculate and cache volume disk usage for all pods
	VolumeStatsAggPeriod metav1.Duration

	// imageMinimumGCAge is the minimum age for an unused image before it is
	// garbage collected.
	ImageMinimumGCAge metav1.Duration
	// imageGCHighThresholdPercent is the percent of disk usage after which
	// image garbage collection is always run. The percent is calculated as
	// this field value out of 100.
	ImageGCHighThresholdPercent int32
	// imageGCLowThresholdPercent is the percent of disk usage before which
	// image garbage collection is never run. Lowest disk usage to garbage
	// collect to. The percent is calculated as this field value out of 100.
	ImageGCLowThresholdPercent int32

	// PodSandboxImage is the image whose network/ipc namespaces
	// containers in each pod will use.
	PodSandboxImage string
}

type Dependencies struct {
	KubeClient clientset.Interface

	Recorder record.EventRecorder
}

// Kubelet is the main kubelet implementation.
type Kubelet struct {

	// Needed to observe and respond to situations that could impact node stability
	evictionManager eviction.Manager

	// Monitor resource usage
	resourceAnalyzer serverstats.ResourceAnalyzer

	// A queue used to trigger pod workers.
	workQueue queue.WorkQueue

	// clock is an interface that provides time related functionality in a way that makes it
	// easy to test the code.
	clock clock.Clock

	// podWorkers handle syncing Pods in response to events.
	podWorkers PodWorkers

	kubeClient clientset.Interface

	// podManager is a facade that abstracts away the various sources of pods
	// this Kubelet services.
	podManager kubepod.Manager

	// Manager for image garbage collection.
	imageManager images.ImageGCManager

	// Container runtime.
	containerRuntime kubecontainer.Runtime
	// Policy for handling garbage collection of dead containers.
	containerGC kubecontainer.GC

	// StatsProvider provides the node and the container stats.
	StatsProvider *stats.Provider
}

func (kubelet *Kubelet) RunOnce() {

}

func NewMainKubelet(
	kubeCfg *KubeletConfiguration,
	kubeDeps *Dependencies,
	nodeName types.NodeName) (*Kubelet, error) {

	klet := &Kubelet{}

	var nodeLister corelisters.NodeLister
	if kubeDeps.KubeClient != nil {
		kubeInformers := informers.NewSharedInformerFactoryWithOptions(kubeDeps.KubeClient, 0, informers.WithTweakListOptions(func(options *metav1.ListOptions) {
			options.FieldSelector = fields.Set{api.ObjectNameField: string(nodeName)}.String()
		}))
		nodeLister = kubeInformers.Core().V1().Nodes().Lister()
	}
	secretManager := secret.NewWatchingSecretManager(kubeDeps.KubeClient)
	configMapManager := configmap.NewWatchingConfigMapManager(kubeDeps.KubeClient)
	// podManager is also responsible for keeping secretManager and configMapManager contents up-to-date.
	mirrorPodClient := kubepod.NewBasicMirrorClient(klet.kubeClient, string(nodeName), nodeLister)
	klet.podManager = kubepod.NewBasicPodManager(mirrorPodClient, secretManager, configMapManager)

	// setup containerGC
	containerGCPolicy := kubecontainer.GCPolicy{
		MinAge:             time.Minute,
		MaxPerPodContainer: 1,
		MaxContainers:      3,
	}
	containerGC, err := kubecontainer.NewContainerGC(klet.containerRuntime, containerGCPolicy, klet.sourcesReady)
	if err != nil {
		return nil, err
	}
	klet.containerGC = containerGC
	// setup imageManager
	imageGCPolicy := images.ImageGCPolicy{
		MinAge:               kubeCfg.ImageMinimumGCAge.Duration,
		HighThresholdPercent: int(kubeCfg.ImageGCHighThresholdPercent),
		LowThresholdPercent:  int(kubeCfg.ImageGCLowThresholdPercent),
	}
	// construct a node reference used for events
	nodeRef := &v1.ObjectReference{
		Kind:      "Node",
		Name:      string(nodeName),
		UID:       types.UID(nodeName),
		Namespace: "",
	}
	imageManager, err := images.NewImageGCManager(klet.containerRuntime, klet.StatsProvider, kubeDeps.Recorder, nodeRef, imageGCPolicy, crOptions.PodSandboxImage)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize image manager: %v", err)
	}
	klet.imageManager = imageManager

	imageFsInfoProvider := cadvisor.NewImageFsInfoProvider("docker", "unix:///var/run/dockershim.sock")
	cAdvisorInterface, err := cadvisor.New(imageFsInfoProvider, "/var/lib/kubelet", cgroupRoots, false)
	if err != nil {
		return nil, err
	}
	klet.StatsProvider = stats.NewCRIStatsProvider(cAdvisorInterface)

	klet.resourceAnalyzer = serverstats.NewResourceAnalyzer(klet.StatsProvider, kubeCfg.VolumeStatsAggPeriod.Duration)
	klet.workQueue = queue.NewBasicWorkQueue(klet.clock)
	klet.podWorkers = newPodWorkers(klet.syncPod, kubeDeps.Recorder, klet.workQueue, klet.resyncInterval, backOffPeriod, klet.podCache)

	thresholds, err := eviction.ParseThresholdConfig(kubeCfg.EnforceNodeAllocatable, kubeCfg.EvictionHard, kubeCfg.EvictionSoft, kubeCfg.EvictionSoftGracePeriod, kubeCfg.EvictionMinimumReclaim)
	if err != nil {
		return nil, err
	}

	evictionConfig := eviction.Config{
		PressureTransitionPeriod: kubeCfg.EvictionPressureTransitionPeriod.Duration,
		MaxPodGracePeriodSeconds: int64(kubeCfg.EvictionMaxPodGracePeriod),
		Thresholds:               thresholds,
		KernelMemcgNotification:  false,
		//PodCgroupRoot:            kubeDeps.ContainerManager.GetPodCgroupRoot(),
		PodCgroupRoot: "/",
	}
	// setup eviction manager
	evictionManager, evictionAdmitHandler := eviction.NewManager(klet.resourceAnalyzer, evictionConfig,
		killPodNow(klet.podWorkers, kubeDeps.Recorder), klet.podManager.GetMirrorPodByPod,
		klet.imageManager, klet.containerGC, kubeDeps.Recorder, nodeRef, klet.clock, etcHostsPathFunc)
	klet.evictionManager = evictionManager
	klet.admitHandlers.AddPodAdmitHandler(evictionAdmitHandler)

	klet.evictionManager.Start(klet.StatsProvider, klet.GetActivePods, klet.podResourcesAreReclaimed, evictionMonitoringPeriod)

	return klet, nil
}
