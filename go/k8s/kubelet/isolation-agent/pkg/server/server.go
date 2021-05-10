package server

import (
	"context"
	"fmt"
	"os"
	"time"

	"k8s-lx1036/k8s/kubelet/isolation-agent/pkg/cgroup"
	"k8s-lx1036/k8s/kubelet/isolation-agent/pkg/scraper"
	topologycpu "k8s-lx1036/k8s/kubelet/isolation-agent/pkg/topology"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	v1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
	api "k8s.io/kubernetes/pkg/apis/core"
	"k8s.io/kubernetes/pkg/kubelet/cadvisor"
	"k8s.io/kubernetes/pkg/kubelet/cm/cpumanager/topology"
	"k8s.io/kubernetes/pkg/kubelet/cm/cpuset"
	kubecontainer "k8s.io/kubernetes/pkg/kubelet/container"
	resourceclient "k8s.io/metrics/pkg/client/clientset/versioned/typed/metrics/v1beta1"
)

type Config struct {
	MetricResolution time.Duration
	ScrapeTimeout    time.Duration

	Kubeconfig string
	Nodename   string

	RemoteRuntimeEndpoint string // "unix:///var/run/dockershim.sock"
	ContainerRuntime      string // "docker"
	ConnectionTimeout     time.Duration
}

type Server struct {

	// The resolution at which metrics-server will retain metrics
	resolution time.Duration
	nodeName   string

	sync      cache.InformerSynced
	informer  informers.SharedInformerFactory
	podLister v1.PodLister

	scraper       *scraper.Scraper
	kubeClient    *kubernetes.Clientset
	cgroupManager *cgroup.Manager
	capacity      corev1.ResourceList
	cpuTopology   *topology.CPUTopology

	remoteRuntimeEndpoint string // "unix:///var/run/dockershim.sock"
	containerRuntime      string // "docker"

	// CgroupRoot is the root cgroup to use for pods.
	// If CgroupsPerQOS is enabled, this is the root of the QoS cgroup hierarchy.
	CgroupRoot string
	// Enable QoS based Cgroup hierarchy: top level cgroups for QoS Classes
	// And all Burstable and BestEffort pods are brought up under their
	// specific top level QoS cgroup.
	CgroupsPerQOS bool
	// driver that the kubelet uses to manipulate cgroups on the host (cgroupfs or systemd)
	CgroupDriver string
}

func NewRestConfig(kubeconfig string) (*rest.Config, error) {
	var config *rest.Config
	if _, err := os.Stat(kubeconfig); err == nil {
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, err
		}
	} else {
		config, err = rest.InClusterConfig()
		if err != nil {
			return nil, err
		}
	}

	return config, nil
}

func NewServer(config *Config) (*Server, error) {
	restConfig, err := NewRestConfig(config.Kubeconfig)
	if err != nil {
		return nil, err
	}
	kubeClient, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("unable to construct lister client: %v", err)
	}

	informer := informers.NewSharedInformerFactoryWithOptions(kubeClient, time.Second*10, informers.WithTweakListOptions(func(options *metav1.ListOptions) {
		options.FieldSelector = fields.Set{api.PodHostField: config.Nodename}.String()
	}))

	metricsClient := resourceclient.NewForConfigOrDie(restConfig)
	//scraper := scraper.NewScraper(metricsClient)
	cgroupManager := cgroup.NewManager(config.RemoteRuntimeEndpoint, config.ConnectionTimeout)

	return &Server{
		//cpuTopology: cpuTopology,
		//capacity: capacity,
		containerRuntime:      config.ContainerRuntime,
		remoteRuntimeEndpoint: config.RemoteRuntimeEndpoint,
		nodeName:              config.Nodename,
		scraper:               scraper.NewScraper(metricsClient),
		kubeClient:            kubeClient,
		cgroupManager:         cgroupManager,
		sync:                  informer.Core().V1().Pods().Informer().HasSynced,
		informer:              informer,
		podLister:             informer.Core().V1().Pods().Lister(),
		resolution:            config.MetricResolution,
	}, nil
}

func (server *Server) getCPUTopology() (*topology.CPUTopology, corev1.ResourceList) {
	rootDirectory := "/var/lib/kubelet"
	imageFsInfoProvider := cadvisor.NewImageFsInfoProvider(server.containerRuntime, server.remoteRuntimeEndpoint)
	cadvisorClient, err := cadvisor.New(imageFsInfoProvider, rootDirectory, server.GetCgroupRoots(),
		cadvisor.UsingLegacyCadvisorStats(server.containerRuntime, server.remoteRuntimeEndpoint))
	if err != nil {
		panic(err)
	}
	machineInfo, err := cadvisorClient.MachineInfo()
	if err != nil {
		panic(err)
	}
	capacity := cadvisor.CapacityFromMachineInfo(machineInfo)
	klog.Info(fmt.Sprintf("cpu: %s, memory: %s", capacity.Cpu().String(), capacity.Memory().String()))
	numaNodeInfo, err := topology.GetNUMANodeInfo()
	if err != nil {
		panic(err)
	}
	cpuTopology, err := topology.Discover(machineInfo, numaNodeInfo)
	if err != nil {
		panic(err)
	}
	allCPUs := cpuTopology.CPUDetails.CPUs()
	klog.Info(fmt.Sprintf("NumCPUs[processor,逻辑核]: %d, NumCores[core,物理核]: %d, NumSockets[NUMA node]: %d, allCPUs: %s",
		cpuTopology.NumCPUs, cpuTopology.NumCores, cpuTopology.NumSockets, allCPUs.String()))

	return cpuTopology, capacity
}

func (server *Server) RunUntil(stopCh <-chan struct{}) error {
	server.informer.Start(stopCh)
	shutdown := cache.WaitForCacheSync(stopCh, server.sync)
	if !shutdown {
		klog.Errorf("can not sync pods in node %s", server.nodeName)
		return nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go wait.UntilWithContext(ctx, func(ctx context.Context) {
		server.tick(ctx, time.Now())
	}, server.resolution)

	<-stopCh

	err := server.RunPreShutdownHooks()
	if err != nil {
		return err
	}

	return nil
}

type ResourceRatio struct {
	cpuRatio    float64
	memoryRatio float64
}

// TODO: 结合在离线pod资源实际使用比率，根据policy计算出离线pod应该独占多少物理核，policy后续在完善先简单写下
func (server *Server) policy(prodResourceRatio, nonProdResourceRatio *ResourceRatio, max int) int {
	numReservedCPUs := 2 // 最小值得是2，即至少一个物理核
	if prodResourceRatio.cpuRatio < 0.2 {
		numReservedCPUs = numReservedCPUs * 2
		if numReservedCPUs > max {
			klog.Warning("")
			numReservedCPUs = 2
		}
	}

	return numReservedCPUs
}

func (server *Server) tick(ctx context.Context, startTime time.Time) {
	pods, err := server.podLister.Pods(metav1.NamespaceAll).List(labels.Everything())
	if err != nil {
		klog.Error(fmt.Sprintf("get pods in node %s err: %v", server.nodeName, err))
		return
	}

	// 每次都重新计算 cpu topology
	cpuTopology, capacity := server.getCPUTopology()
	// 由于可能会超卖，所以不要使用 node .status.allocatable
	prodMetrics, nonProdMetrics, err := server.scraper.Scrape(ctx, pods)
	if err != nil {
		klog.Error(err)
	}
	prodCpuRatio := float64(prodMetrics.Cpu().Value()) / float64(capacity.Cpu().Value())
	nonProdCpuRatio := float64(nonProdMetrics.Cpu().Value()) / float64(capacity.Cpu().Value())
	prodMemoryRatio := float64(prodMetrics.Memory().Value()) / float64(capacity.Memory().Value())
	nonProdMemoryRatio := float64(nonProdMetrics.Memory().Value()) / float64(capacity.Memory().Value())
	klog.Info(prodCpuRatio, nonProdCpuRatio, prodMemoryRatio, nonProdMemoryRatio)
	prodResourceRatio := &ResourceRatio{
		cpuRatio:    prodCpuRatio,
		memoryRatio: prodMemoryRatio,
	}
	nonProdResourceRatio := &ResourceRatio{
		cpuRatio:    nonProdCpuRatio,
		memoryRatio: nonProdMemoryRatio,
	}

	allCPUs := cpuTopology.CPUDetails.CPUs()
	numReservedCPUs := server.policy(prodResourceRatio, nonProdResourceRatio, allCPUs.Size())
	reserved, err := topologycpu.TakeCPUByTopology(cpuTopology, allCPUs, numReservedCPUs)
	if err != nil {
		klog.Error(err)
		return
	}

	prodCPUSet := allCPUs.Difference(reserved)
	klog.Info(fmt.Sprintf("take cpuset %s for nonProdPod, cpuset %s for ProdPod", reserved.String(), prodCPUSet.String()))

	// default all cpus
	cpuSet := allCPUs.Clone()
	// TODO: get cpuset.CPUSet
	for _, pod := range GetActivePods(pods) {
		if IsProdPod(pod) {
			cpuSet = prodCPUSet.Clone()
		} else if IsNonProdPod(pod) {
			cpuSet = reserved.Clone()
		}

		podStatus := pod.Status
		allContainers := pod.Spec.InitContainers
		allContainers = append(allContainers, pod.Spec.Containers...)
		for _, container := range allContainers {
			containerID, err := findContainerIDByName(&podStatus, container.Name)
			if err != nil {
				continue
			}

			// filter container by container status
			containerStatus, err := findContainerStatusByName(&podStatus, container.Name)
			if err != nil {
				continue
			}
			if containerStatus.State.Waiting != nil ||
				(containerStatus.State.Waiting == nil && containerStatus.State.Running == nil && containerStatus.State.Terminated == nil) {
				klog.Warningf("reconcileState: skipping container; container still in the waiting state (pod: %s, container: %s, error: %v)",
					pod.Name, container.Name, err)
				continue
			}
			if containerStatus.State.Terminated != nil {
				klog.Warningf("[cpumanager] reconcileState: ignoring terminated container (pod: %s, container id: %s)",
					pod.Name, containerID)
				continue
			}

			klog.Info(fmt.Sprintf("[cpumanager] reconcileState: updating container (pod: %s, container: %s, container id: %s, cpuset: %s)",
				pod.Name, container.Name, containerID, cpuSet.String()))
			err = server.updateContainerCPUSet(containerID, cpuSet)
			if err != nil {
				klog.Error(fmt.Sprintf("[cpumanager] reconcileState: failed to update container (pod: %s, container: %s, container id: %s, cpuset: %s, error: %v)",
					pod.Name, container.Name, containerID, cpuSet.String(), err))
				continue
			}
		}
	}
}

func (server *Server) updateContainerCPUSet(containerID string, cpus cpuset.CPUSet) error {
	return server.cgroupManager.UpdateContainerResource(containerID, cpus)
}

func (server *Server) RunPreShutdownHooks() error {
	return nil
}

// INFO: @see pkg/kubelet/cm/cpumanager/cpu_manager.go::findContainerIDByName()
func findContainerIDByName(status *corev1.PodStatus, name string) (string, error) {
	allStatuses := status.InitContainerStatuses
	allStatuses = append(allStatuses, status.ContainerStatuses...)
	for _, container := range allStatuses {
		if container.Name == name && container.ContainerID != "" {
			cid := &kubecontainer.ContainerID{}
			err := cid.ParseString(container.ContainerID)
			if err != nil {
				return "", err
			}
			return cid.ID, nil
		}
	}

	return "", fmt.Errorf("unable to find ID for container with name %v in pod status (it may not be running)", name)
}

func findContainerStatusByName(status *corev1.PodStatus, name string) (*corev1.ContainerStatus, error) {
	for _, status := range append(status.InitContainerStatuses, status.ContainerStatuses...) {
		if status.Name == name {
			return &status, nil
		}
	}

	return nil, fmt.Errorf("unable to find status for container with name %v in pod status (it may not be running)", name)
}
