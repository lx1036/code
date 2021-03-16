package stats

import (
	"fmt"
	"sync"

	statsapi "k8s-lx1036/k8s/kubelet/pkg/apis/v1alpha1"
	"k8s-lx1036/k8s/kubelet/pkg/cadvisor"
	"k8s-lx1036/k8s/kubelet/pkg/server/stats"

	cadvisorapiv2 "github.com/google/cadvisor/info/v2"
	internalapi "k8s.io/cri-api/pkg/apis"

	"k8s.io/apimachinery/pkg/types"
	runtimeapi "k8s.io/cri-api/pkg/apis/runtime/v1alpha2"
	"k8s.io/klog/v2"
	kubecontainer "k8s.io/kubernetes/pkg/kubelet/container"
)

// cpuUsageRecord holds the cpu usage stats and the calculated usageNanoCores.
type cpuUsageRecord struct {
	stats          *runtimeapi.CpuUsage
	usageNanoCores *uint64
}

// criStatsProvider implements the containerStatsProvider interface by getting
// the container stats from CRI.
type criStatsProvider struct {
	// cadvisor is used to get the node root filesystem's stats (such as the
	// capacity/available bytes/inodes) that will be populated in per container
	// filesystem stats.
	cadvisor cadvisor.Interface
	// resourceAnalyzer is used to get the volume stats of the pods.
	resourceAnalyzer stats.ResourceAnalyzer
	// runtimeService is used to get the status and stats of the pods and its
	// managed containers.
	runtimeService internalapi.RuntimeService
	// imageService is used to get the stats of the image filesystem.
	imageService internalapi.ImageManagerService
	// logMetrics provides the metrics for container logs
	logMetricsService LogMetricsService
	// osInterface is the interface for syscalls.
	osInterface kubecontainer.OSInterface

	// cpuUsageCache caches the cpu usage for containers.
	cpuUsageCache map[string]*cpuUsageRecord
	mutex         sync.RWMutex
}

func (p *criStatsProvider) listPodStats(updateCPUNanoCoreUsage bool) ([]statsapi.PodStats, error) {
	// Gets node root filesystem information, which will be used to populate
	// the available and capacity bytes/inodes in container stats.
	rootFsInfo, err := p.cadvisor.RootFsInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to get rootFs info: %v", err)
	}

	containers, err := p.runtimeService.ListContainers(&runtimeapi.ContainerFilter{})
	if err != nil {
		return nil, fmt.Errorf("failed to list all containers: %v", err)
	}

	// Creates pod sandbox map.
	podSandboxMap := make(map[string]*runtimeapi.PodSandbox)
	podSandboxes, err := p.runtimeService.ListPodSandbox(&runtimeapi.PodSandboxFilter{})
	if err != nil {
		return nil, fmt.Errorf("failed to list all pod sandboxes: %v", err)
	}
	podSandboxes = removeTerminatedPods(podSandboxes)
	for _, s := range podSandboxes {
		podSandboxMap[s.Id] = s
	}
	// fsIDtoInfo is a map from filesystem id to its stats. This will be used
	// as a cache to avoid querying cAdvisor for the filesystem stats with the
	// same filesystem id many times.
	fsIDtoInfo := make(map[runtimeapi.FilesystemIdentifier]*cadvisorapiv2.FsInfo)

	// sandboxIDToPodStats is a temporary map from sandbox ID to its pod stats.
	sandboxIDToPodStats := make(map[string]*statsapi.PodStats)

	resp, err := p.runtimeService.ListContainerStats(&runtimeapi.ContainerStatsFilter{})
	if err != nil {
		return nil, fmt.Errorf("failed to list all container stats: %v", err)
	}

	containers = removeTerminatedContainers(containers)
	// Creates container map.
	containerMap := make(map[string]*runtimeapi.Container)
	for _, c := range containers {
		containerMap[c.Id] = c
	}

	allInfos, err := getCadvisorContainerInfo(p.cadvisor)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch cadvisor stats: %v", err)
	}
	caInfos := getCRICadvisorStats(allInfos)

	// get network stats for containers.
	// This is only used on Windows. For other platforms, (nil, nil) should be returned.
	containerNetworkStats, err := p.listContainerNetworkStats()
	if err != nil {
		return nil, fmt.Errorf("failed to list container network stats: %v", err)
	}

	for _, stats := range resp {
		containerID := stats.Attributes.Id
		container, found := containerMap[containerID]
		if !found {
			continue
		}

		podSandboxID := container.PodSandboxId
		podSandbox, found := podSandboxMap[podSandboxID]
		if !found {
			continue
		}

		// Creates the stats of the pod (if not created yet) which the
		// container belongs to.
		ps, found := sandboxIDToPodStats[podSandboxID]
		if !found {
			ps = buildPodStats(podSandbox)
			sandboxIDToPodStats[podSandboxID] = ps
		}

		// Fill available stats for full set of required pod stats
		cs := p.makeContainerStats(stats, container, &rootFsInfo, fsIDtoInfo, podSandbox.GetMetadata(), updateCPUNanoCoreUsage)
		p.addPodNetworkStats(ps, podSandboxID, caInfos, cs, containerNetworkStats[podSandboxID])
		p.addPodCPUMemoryStats(ps, types.UID(podSandbox.Metadata.Uid), allInfos, cs)
		p.addProcessStats(ps, types.UID(podSandbox.Metadata.Uid), allInfos, cs)

		// If cadvisor stats is available for the container, use it to populate
		// container stats
		caStats, caFound := caInfos[containerID]
		if !caFound {
			klog.V(5).Infof("Unable to find cadvisor stats for %q", containerID)
		} else {
			p.addCadvisorContainerStats(cs, &caStats)
		}
		ps.Containers = append(ps.Containers, *cs)
	}
	// cleanup outdated caches.
	p.cleanupOutdatedCaches()

	result := make([]statsapi.PodStats, 0, len(sandboxIDToPodStats))
	for _, s := range sandboxIDToPodStats {
		p.makePodStorageStats(s, &rootFsInfo)
		result = append(result, *s)
	}
	return result, nil
}

// ListPodStats returns the stats of all the pod-managed containers.
func (p *criStatsProvider) ListPodStats() ([]statsapi.PodStats, error) {
	// Don't update CPU nano core usage.
	return p.listPodStats(false)
}

// ListPodStatsAndUpdateCPUNanoCoreUsage updates the cpu nano core usage for
// the containers and returns the stats for all the pod-managed containers.
// This is a workaround because CRI runtimes do not supply nano core usages,
// so this function calculate the difference between the current and the last
// (cached) cpu stats to calculate this metrics. The implementation assumes a
// single caller to periodically invoke this function to update the metrics. If
// there exist multiple callers, the period used to compute the cpu usage may
// vary and the usage could be incoherent (e.g., spiky). If no caller calls
// this function, the cpu usage will stay nil. Right now, eviction manager is
// the only caller, and it calls this function every 10s.
func (p *criStatsProvider) ListPodStatsAndUpdateCPUNanoCoreUsage() ([]statsapi.PodStats, error) {
	return p.listPodStats(true)
}

func (p *criStatsProvider) ListPodCPUAndMemoryStats() ([]statsapi.PodStats, error) {
	panic("implement me")
}

func (p *criStatsProvider) ImageFsStats() (*statsapi.FsStats, error) {
	panic("implement me")
}

func (p *criStatsProvider) ImageFsDevice() (string, error) {
	panic("implement me")
}

// newCRIStatsProvider returns a containerStatsProvider implementation that
// provides container stats using CRI.
func newCRIStatsProvider(
	cadvisor cadvisor.Interface,
	resourceAnalyzer stats.ResourceAnalyzer,
	runtimeService internalapi.RuntimeService,
	imageService internalapi.ImageManagerService,
	logMetricsService LogMetricsService,
	osInterface kubecontainer.OSInterface,
) containerStatsProvider {
	return &criStatsProvider{
		cadvisor:          cadvisor,
		resourceAnalyzer:  resourceAnalyzer,
		runtimeService:    runtimeService,
		imageService:      imageService,
		logMetricsService: logMetricsService,
		osInterface:       osInterface,
		cpuUsageCache:     make(map[string]*cpuUsageRecord),
	}
}
