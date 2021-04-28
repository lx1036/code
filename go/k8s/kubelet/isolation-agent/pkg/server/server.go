package server

import (
	"context"
	"fmt"
	"os"
	"time"

	"k8s-lx1036/k8s/kubelet/isolation-agent/pkg/cgroup"
	"k8s-lx1036/k8s/kubelet/isolation-agent/pkg/policy"
	"k8s-lx1036/k8s/kubelet/isolation-agent/pkg/scraper"

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
	ConnectionTimeout     time.Duration
}

// server scrapes metrics and serves then using k8s api.
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
}

func NewRestConfig(kubeconfig string) (*rest.Config, error) {
	var config *rest.Config
	if _, err := os.Stat(kubeconfig); err == nil {
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, err
		}
	} else { //Use Incluster Configuration
		config, err = rest.InClusterConfig()
		if err != nil {
			return nil, err
		}
	}

	// Use protobufs for communication with apiserver
	//config.ContentType = "application/vnd.kubernetes.protobuf"
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
		nodeName:      config.Nodename,
		scraper:       scraper.NewScraper(metricsClient),
		kubeClient:    kubeClient,
		cgroupManager: cgroupManager,
		sync:          informer.Core().V1().Pods().Informer().HasSynced,
		informer:      informer,
		podLister:     informer.Core().V1().Pods().Lister(),
		resolution:    config.MetricResolution,
	}, nil

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

func (server *Server) tick(ctx context.Context, startTime time.Time) {
	pods, err := server.podLister.Pods(metav1.NamespaceAll).List(labels.Everything())
	if err != nil {
		klog.Error(fmt.Sprintf("get pods in node %s err: %v", server.nodeName, err))
		return
	}

	prodMetrics, nonProdMetrics, err := server.scraper.Scrape(ctx, pods)
	if err != nil {
		klog.Error(err)
	}

	currentNode, err := server.kubeClient.CoreV1().Nodes().Get(context.TODO(), server.nodeName, metav1.GetOptions{})
	if err != nil {
		klog.Error(err)
		return
	}

	prodCpu := prodMetrics[corev1.ResourceCPU]
	nonProdCpu := nonProdMetrics[corev1.ResourceCPU]
	prodMemory := prodMetrics[corev1.ResourceMemory]
	nonProdMemory := nonProdMetrics[corev1.ResourceMemory]
	nodeCpuResource := currentNode.Status.Allocatable[corev1.ResourceCPU]
	nodeMemoryResource := currentNode.Status.Allocatable[corev1.ResourceMemory]
	prodCpuRatio := float64(prodCpu.Value()) / float64(nodeCpuResource.Value())
	nonProdCpuRatio := float64(nonProdCpu.Value()) / float64(nodeCpuResource.Value())
	prodMemoryRatio := float64(prodMemory.Value()) / float64(nodeMemoryResource.Value())
	nonProdMemoryRatio := float64(nonProdMemory.Value()) / float64(nodeMemoryResource.Value())

	klog.Info(prodCpuRatio, nonProdCpuRatio, prodMemoryRatio, nonProdMemoryRatio)

	// TODO: get cpuset.CPUSet
	for _, pod := range pods {
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
				klog.Warningf("reconcileState: skipping container; container still in the waiting state (pod: %s, container: %s, error: %v)", pod.Name, container.Name, err)
				//failure = append(failure, reconciledContainer{pod.Name, container.Name, ""})
				continue
			}
			if containerStatus.State.Terminated != nil {
				klog.Warningf("[cpumanager] reconcileState: ignoring terminated container (pod: %s, container id: %s)", pod.Name, containerID)
				continue
			}

			cpuSet := policy.GetCPUSetOrDefault(prodCpuRatio)

			klog.V(4).Infof("[cpumanager] reconcileState: updating container (pod: %s, container: %s, container id: %s, cpuset: \"%v\")", pod.Name, container.Name, containerID, cset)
			err = server.updateContainerCPUSet(containerID, cpuSet)
			if err != nil {
				klog.Errorf("[cpumanager] reconcileState: failed to update container (pod: %s, container: %s, container id: %s, cpuset: \"%v\", error: %v)", pod.Name, container.Name, containerID, cset, err)

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
