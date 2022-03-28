package server

import (
	"context"
	"fmt"
	"time"

	"k8s-lx1036/k8s/kubelet/isolation-agent/cmd/app/options"
	"k8s-lx1036/k8s/kubelet/isolation-agent/pkg/cgroup"
	"k8s-lx1036/k8s/kubelet/isolation-agent/pkg/scraper"
	topologycpu "k8s-lx1036/k8s/kubelet/isolation-agent/pkg/topology"
	"k8s-lx1036/k8s/kubelet/isolation-agent/pkg/utils"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	v1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
	coreapi "k8s.io/kubernetes/pkg/apis/core"
	"k8s.io/kubernetes/pkg/kubelet/cadvisor"
	"k8s.io/kubernetes/pkg/kubelet/cm"
	"k8s.io/kubernetes/pkg/kubelet/cm/cpumanager/topology"
	"k8s.io/kubernetes/pkg/kubelet/cm/cpuset"
	resourceclient "k8s.io/metrics/pkg/client/clientset/versioned/typed/metrics/v1beta1"
)

type Server struct {
	podLister   v1.PodLister
	podInformer cache.SharedIndexInformer

	scraper       *scraper.Scraper
	kubeClient    *kubernetes.Clientset
	cgroupManager *cgroup.Manager

	option *options.Options
}

// TODO: add prometheus metrics
func NewServer(option *options.Options) (*Server, error) {
	restConfig, err := clientcmd.BuildConfigFromFlags("", option.Kubeconfig)
	if err != nil {
		return nil, err
	}
	kubeClient, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("unable to construct lister client: %v", err)
	}

	informerFactory := informers.NewSharedInformerFactoryWithOptions(kubeClient, time.Second*10, informers.WithTweakListOptions(func(options *metav1.ListOptions) {
		options.FieldSelector = fields.Set{coreapi.PodHostField: option.Nodename}.String()
	}))

	podInformer := informerFactory.Core().V1().Pods().Informer()
	metricsClient := resourceclient.NewForConfigOrDie(restConfig)
	cgroupManager, err := cgroup.NewManager(option.RemoteRuntimeEndpoint, option.RuntimeRequestTimeout)
	if err != nil {
		return nil, err
	}

	return &Server{
		option:        option,
		scraper:       scraper.NewScraper(metricsClient),
		cgroupManager: cgroupManager,
		podInformer:   podInformer,
		podLister:     informerFactory.Core().V1().Pods().Lister(),
	}, nil
}

// INFO: @see https://github.com/kubernetes/kubernetes/blob/release-1.19/cmd/kubelet/app/server.go#L604-L606
func (server *Server) GetCgroupRoots() []string {
	var cgroupRoots []string
	nodeAllocatableRoot := cm.NodeAllocatableRoot(server.option.CgroupRoot, server.option.CgroupsPerQOS, server.option.CgroupDriver)
	cgroupRoots = append(cgroupRoots, nodeAllocatableRoot) // "/kubepods"

	return cgroupRoots
}

func (server *Server) getCPUTopology() (*topology.CPUTopology, corev1.ResourceList, error) {
	imageFsInfoProvider := cadvisor.NewImageFsInfoProvider(server.option.ContainerRuntime, server.option.RemoteRuntimeEndpoint)
	cadvisorClient, err := cadvisor.New(imageFsInfoProvider, server.option.RootDirectory, server.GetCgroupRoots(),
		cadvisor.UsingLegacyCadvisorStats(server.option.ContainerRuntime, server.option.RemoteRuntimeEndpoint))
	if err != nil {
		return nil, nil, err
	}
	machineInfo, err := cadvisorClient.MachineInfo()
	if err != nil {
		return nil, nil, err
	}
	capacity := cadvisor.CapacityFromMachineInfo(machineInfo)
	klog.Info(fmt.Sprintf("[node capacity]node %s has resources cpu: %s, memory: %s", server.option.Nodename, capacity.Cpu().String(), capacity.Memory().String()))
	numaNodeInfo, err := topology.GetNUMANodeInfo()
	if err != nil {
		return nil, nil, err
	}
	cpuTopology, err := topology.Discover(machineInfo, numaNodeInfo)
	if err != nil {
		return nil, nil, err
	}
	allCPUs := cpuTopology.CPUDetails.CPUs()
	klog.Info(fmt.Sprintf("[node capacity]node %s has cpu topology: %d processors, %d cores, %d sockets, all CPU ID: %s",
		server.option.Nodename, cpuTopology.NumCPUs, cpuTopology.NumCores, cpuTopology.NumSockets, allCPUs.String()))

	return cpuTopology, capacity, nil
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
			klog.Warning(fmt.Sprintf("[warning]numReservedCPUs %d is above max %d", numReservedCPUs, max))
			numReservedCPUs = 2
		}
	}

	return numReservedCPUs
}

func (server *Server) tick(ctx context.Context, startTime time.Time) {
	pods, err := server.podLister.Pods(metav1.NamespaceAll).List(labels.Everything())
	if err != nil {
		klog.Error(fmt.Sprintf("get pods in node %s err: %v", server.option.Nodename, err))
		return
	}
	activePods := GetActivePods(pods)

	klog.Info(fmt.Sprintf("[1]calculate cpu topology..."))
	// 每次都重新计算 cpu topology
	cpuTopology, capacity, err := server.getCPUTopology()
	if err != nil {
		klog.Error(fmt.Sprintf("get cpu topology err: %v", err))
		return
	}

	klog.Info(fmt.Sprintf("[2]calculate pods resource metrics..."))
	// 由于可能会超卖，所以不要使用 Node.status.allocatable
	prodMetrics, nonProdMetrics, err := server.scraper.Scrape(ctx, activePods)
	if err != nil {
		klog.Error(fmt.Sprintf("get pod resource err: %v", err))
		// no return
	}
	prodCpuRatio := float64(prodMetrics.Cpu().Value()) / float64(capacity.Cpu().Value())
	nonProdCpuRatio := float64(nonProdMetrics.Cpu().Value()) / float64(capacity.Cpu().Value())
	prodMemoryRatio := float64(prodMetrics.Memory().Value()) / float64(capacity.Memory().Value())
	nonProdMemoryRatio := float64(nonProdMetrics.Memory().Value()) / float64(capacity.Memory().Value())
	klog.Info(fmt.Sprintf("[ratio]prod pod ratio: cpu %f, memory %f; non prod pod ratio: cpu %f, memory %f.", prodCpuRatio, nonProdCpuRatio, prodMemoryRatio, nonProdMemoryRatio))
	prodResourceRatio := &ResourceRatio{
		cpuRatio:    prodCpuRatio,
		memoryRatio: prodMemoryRatio,
	}
	nonProdResourceRatio := &ResourceRatio{
		cpuRatio:    nonProdCpuRatio,
		memoryRatio: nonProdMemoryRatio,
	}

	klog.Info(fmt.Sprintf("[3]calculate pods cpuset..."))
	allCPUs := cpuTopology.CPUDetails.CPUs()
	numReservedCPUs := server.policy(prodResourceRatio, nonProdResourceRatio, allCPUs.Size())
	klog.Info(fmt.Sprintf("[policy]need %d processors at time %s", numReservedCPUs, startTime.String()))
	reserved, err := topologycpu.TakeCPUByTopology(cpuTopology, allCPUs, numReservedCPUs)
	if err != nil {
		klog.Error(err)
		return
	}
	prodCPUSet := allCPUs.Difference(reserved)
	klog.Info(fmt.Sprintf("[policy]take cpuset %s for nonProdPod, cpuset %s for ProdPod", reserved.String(), prodCPUSet.String()))

	klog.Info(fmt.Sprintf("[4]update containers cpuset..."))
	// default all cpus
	cpuSet := allCPUs.Clone()
	for _, pod := range activePods {
		if utils.IsProdPod(pod) {
			cpuSet = prodCPUSet.Clone()
		} else if utils.IsNonProdPod(pod) {
			cpuSet = reserved.Clone()
		}

		podStatus := pod.Status
		allContainers := pod.Spec.InitContainers
		allContainers = append(allContainers, pod.Spec.Containers...)
		for _, container := range allContainers {
			containerID, err := findContainerIDByName(&podStatus, container.Name)
			if err != nil {
				klog.Warning(fmt.Sprintf("[container cpuset]ignoring container for container id(pod: %s, container: %s, err: %v)", container.Name, pod.Name, err))
				continue
			}

			// filter container by container status
			containerStatus, err := findContainerStatusByName(&podStatus, container.Name)
			if err != nil {
				klog.Warning(fmt.Sprintf("[container cpuset]ignoring container for container status(pod: %s, container: %s, err: %v)", container.Name, pod.Name, err))
				continue
			}
			if containerStatus.State.Waiting != nil ||
				(containerStatus.State.Waiting == nil && containerStatus.State.Running == nil && containerStatus.State.Terminated == nil) {
				klog.Warning(fmt.Sprintf("[container cpuset]ignoring waiting state container (pod: %s, container: %s)",
					pod.Name, container.Name))
				continue
			}
			if containerStatus.State.Terminated != nil {
				klog.Warning(fmt.Sprintf("[container cpuset]ignoring terminated container (pod: %s, container id: %s)",
					pod.Name, containerID))
				continue
			}

			klog.Info(fmt.Sprintf("[container cpuset]updating container cpuset(pod name: %s, pod uid: %s, container name: %s, container id: %s, cpuset: %s)",
				pod.Name, pod.UID, container.Name, containerID, cpuSet.String()))

			// if debug, skip updating container cpuset
			if server.option.Debug {
				continue
			}
			err = server.updateContainerCPUSet(containerID, cpuSet)
			if err != nil {
				klog.Error(fmt.Sprintf("[container cpuset]failed to update container (pod name: %s, pod uid: %s, container name: %s, container id: %s, cpuset: %s, error: %v)",
					pod.Name, pod.UID, container.Name, containerID, cpuSet.String(), err))
				continue
			}
		}
	}

	latency := time.Since(startTime).Seconds()
	klog.Info(fmt.Sprintf("[5]consume %f seconds...", latency))
}

func (server *Server) updateContainerCPUSet(containerID string, cpus cpuset.CPUSet) error {
	return server.cgroupManager.UpdateContainerResource(containerID, cpus)
}

func (server *Server) RunPreShutdownHooks() error {
	// do something
	return nil
}

func (server *Server) RunUntil(stopCh <-chan struct{}) error {
	server.podInformer.Run(stopCh)
	shutdown := cache.WaitForCacheSync(stopCh, server.podInformer.HasSynced)
	if !shutdown {
		klog.Errorf("can not sync pods in node %s", server.option.Nodename)
		return nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	klog.Info(fmt.Sprintf("updating container cpuset..."))
	go wait.UntilWithContext(ctx, func(ctx context.Context) {
		server.tick(ctx, time.Now())
	}, server.option.ReconcilePeriod)

	<-stopCh

	err := server.RunPreShutdownHooks()
	if err != nil {
		return err
	}

	return nil
}
