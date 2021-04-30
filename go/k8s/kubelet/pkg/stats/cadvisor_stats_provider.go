package stats

import (
	"fmt"
	"path"
	"sort"
	"strings"

	statsapi "k8s-lx1036/k8s/kubelet/pkg/apis/stats/v1alpha1"
	"k8s-lx1036/k8s/kubelet/pkg/cadvisor"
	cadvisorapiv2 "k8s-lx1036/k8s/kubelet/pkg/cadvisor/pkg/info/v2"
	"k8s-lx1036/k8s/kubelet/pkg/cm"
	kubecontainer "k8s-lx1036/k8s/kubelet/pkg/container"
	"k8s-lx1036/k8s/kubelet/pkg/leaky"
	"k8s-lx1036/k8s/kubelet/pkg/server/stats"
	"k8s-lx1036/k8s/kubelet/pkg/status"
	kubetypes "k8s-lx1036/k8s/kubelet/pkg/types"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
)

// cadvisorStatsProvider implements the containerStatsProvider interface by
// getting the container stats from cAdvisor. This is needed by
// integrations which do not provide stats from CRI. See
// `pkg/kubelet/cadvisor/util.go#UsingLegacyCadvisorStats` for the logic for
// determining which integrations do not provide stats from CRI.
type cadvisorStatsProvider struct {
	// cadvisor is used to get the stats of the cgroup for the containers that
	// are managed by pods.
	cadvisor cadvisor.Interface
	// resourceAnalyzer is used to get the volume stats of the pods.
	resourceAnalyzer stats.ResourceAnalyzer
	// imageService is used to get the stats of the image filesystem.
	imageService kubecontainer.ImageService
	// statusProvider is used to get pod metadata
	statusProvider status.PodStatusProvider
}

func (p *cadvisorStatsProvider) ListPodStatsAndUpdateCPUNanoCoreUsage() ([]statsapi.PodStats, error) {
	panic("implement me")
}

func (p *cadvisorStatsProvider) ListPodCPUAndMemoryStats() ([]statsapi.PodStats, error) {
	panic("implement me")
}

func (p *cadvisorStatsProvider) ImageFsStats() (*statsapi.FsStats, error) {
	panic("implement me")
}

func (p *cadvisorStatsProvider) ImageFsDevice() (string, error) {
	panic("implement me")
}

// ListPodStats returns the stats of all the pod-managed containers.
func (p *cadvisorStatsProvider) ListPodStats() ([]statsapi.PodStats, error) {
	// Gets node root filesystem information and image filesystem stats, which
	// will be used to populate the available and capacity bytes/inodes in
	// container stats.
	rootFsInfo, err := p.cadvisor.RootFsInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to get rootFs info: %v", err)
	}
	imageFsInfo, err := p.cadvisor.ImagesFsInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to get imageFs info: %v", err)
	}
	// INFO: 这里是关键，已经通过 cadvisor 获得所有容器的资源实际使用量
	infos, err := getCadvisorContainerInfo(p.cadvisor)
	if err != nil {
		return nil, fmt.Errorf("failed to get container info from cadvisor: %v", err)
	}

	// removeTerminatedContainerInfo will also remove pod level cgroups, so save the infos into allInfos first
	allInfos := infos
	infos = removeTerminatedContainerInfo(infos)
	// Map each container to a pod and update the PodStats with container data.
	podToStats := map[statsapi.PodReference]*statsapi.PodStats{}
	for key, cinfo := range infos {
		// On systemd using devicemapper each mount into the container has an
		// associated cgroup. We ignore them to ensure we do not get duplicate
		// entries in our summary. For details on .mount units:
		// http://man7.org/linux/man-pages/man5/systemd.mount.5.html
		if strings.HasSuffix(key, ".mount") {
			continue
		}
		// Build the Pod key if this container is managed by a Pod
		if !isPodManagedContainer(&cinfo) {
			continue
		}
		ref := buildPodRef(cinfo.Spec.Labels)

		// Lookup the PodStats for the pod using the PodRef. If none exists,
		// initialize a new entry.
		podStats, found := podToStats[ref]
		if !found {
			podStats = &statsapi.PodStats{PodRef: ref}
			podToStats[ref] = podStats
		}

		// Update the PodStats entry with the stats from the container by
		// adding it to podStats.Containers.
		containerName := kubetypes.GetContainerName(cinfo.Spec.Labels)
		if containerName == leaky.PodInfraContainerName {
			// Special case for infrastructure container which is hidden from
			// the user and has network stats.
			podStats.Network = cadvisorInfoToNetworkStats(&cinfo)
		} else {
			podStats.Containers = append(podStats.Containers, *cadvisorInfoToContainerStats(containerName, &cinfo, &rootFsInfo, &imageFsInfo))
		}
	}

	// Add each PodStats to the result.
	result := make([]statsapi.PodStats, 0, len(podToStats))
	for _, podStats := range podToStats {
		// Lookup the volume stats for each pod.
		podUID := types.UID(podStats.PodRef.UID)
		var ephemeralStats []statsapi.VolumeStats
		if vstats, found := p.resourceAnalyzer.GetPodVolumeStats(podUID); found {
			ephemeralStats = make([]statsapi.VolumeStats, len(vstats.EphemeralVolumes))
			copy(ephemeralStats, vstats.EphemeralVolumes)
			podStats.VolumeStats = append(append([]statsapi.VolumeStats{}, vstats.EphemeralVolumes...), vstats.PersistentVolumes...)
		}
		podStats.EphemeralStorage = calcEphemeralStorage(podStats.Containers, ephemeralStats, &rootFsInfo, nil, false)
		// Lookup the pod-level cgroup's CPU and memory stats
		podInfo := getCadvisorPodInfoFromPodUID(podUID, allInfos)
		if podInfo != nil {
			cpu, memory := cadvisorInfoToCPUandMemoryStats(podInfo)
			podStats.CPU = cpu
			podStats.Memory = memory
			podStats.ProcessStats = cadvisorInfoToProcessStats(podInfo)
		}

		podStatus, found := p.statusProvider.GetPodStatus(podUID)
		if found && podStatus.StartTime != nil && !podStatus.StartTime.IsZero() {
			podStats.StartTime = *podStatus.StartTime
			// only append stats if we were able to get the start time of the pod
			result = append(result, *podStats)
		}
	}

	return result, nil
}

// getCadvisorPodInfoFromPodUID returns a pod cgroup information by matching the podUID with its CgroupName identifier base name
func getCadvisorPodInfoFromPodUID(podUID types.UID, infos map[string]cadvisorapiv2.ContainerInfo) *cadvisorapiv2.ContainerInfo {
	for key, info := range infos {
		key = path.Base(key)
		if cm.GetPodCgroupNameSuffix(podUID) == key {
			return &info
		}
	}
	return nil
}

// removeTerminatedContainerInfo returns the specified containerInfo but with
// the stats of the terminated containers removed.
//
// A ContainerInfo is considered to be of a terminated container if it has an
// older CreationTime and zero CPU instantaneous and memory RSS usage.
func removeTerminatedContainerInfo(containerInfo map[string]cadvisorapiv2.ContainerInfo) map[string]cadvisorapiv2.ContainerInfo {
	cinfoMap := make(map[containerID][]containerInfoWithCgroup)
	for key, cinfo := range containerInfo {
		if !isPodManagedContainer(&cinfo) {
			continue
		}
		cinfoID := containerID{
			podRef:        buildPodRef(cinfo.Spec.Labels),
			containerName: kubetypes.GetContainerName(cinfo.Spec.Labels),
		}
		cinfoMap[cinfoID] = append(cinfoMap[cinfoID], containerInfoWithCgroup{
			cinfo:  cinfo,
			cgroup: key,
		})
	}
	result := make(map[string]cadvisorapiv2.ContainerInfo)
	for _, refs := range cinfoMap {
		if len(refs) == 1 {
			result[refs[0].cgroup] = refs[0].cinfo
			continue
		}
		sort.Sort(ByCreationTime(refs))
		for i := len(refs) - 1; i >= 0; i-- {
			if hasMemoryAndCPUInstUsage(&refs[i].cinfo) {
				result[refs[i].cgroup] = refs[i].cinfo
				break
			}
		}
	}
	return result
}

// newCadvisorStatsProvider returns a containerStatsProvider that provides
// container stats from cAdvisor.
func newCadvisorStatsProvider(
	cadvisor cadvisor.Interface,
	resourceAnalyzer stats.ResourceAnalyzer,
	imageService kubecontainer.ImageService,
	statusProvider status.PodStatusProvider,
) containerStatsProvider {
	return &cadvisorStatsProvider{
		cadvisor:         cadvisor,
		resourceAnalyzer: resourceAnalyzer,
		imageService:     imageService,
		statusProvider:   statusProvider,
	}
}

func getCadvisorContainerInfo(ca cadvisor.Interface) (map[string]cadvisorapiv2.ContainerInfo, error) {
	infos, err := ca.ContainerInfoV2("/", cadvisorapiv2.RequestOptions{
		IdType:    cadvisorapiv2.TypeName,
		Count:     2, // 2 samples are needed to compute "instantaneous" CPU
		Recursive: true,
	})
	if err != nil {
		if _, ok := infos["/"]; ok {
			// If the failure is partial, log it and return a best-effort
			// response.
			klog.Errorf("Partial failure issuing cadvisor.ContainerInfoV2: %v", err)
		} else {
			return nil, fmt.Errorf("failed to get root cgroup stats: %v", err)
		}
	}

	return infos, nil
}

// ByCreationTime implements sort.Interface for []containerInfoWithCgroup based
// on the cinfo.Spec.CreationTime field.
type ByCreationTime []containerInfoWithCgroup

func (a ByCreationTime) Len() int      { return len(a) }
func (a ByCreationTime) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByCreationTime) Less(i, j int) bool {
	if a[i].cinfo.Spec.CreationTime.Equal(a[j].cinfo.Spec.CreationTime) {
		// There shouldn't be two containers with the same name and/or the same
		// creation time. However, to make the logic here robust, we break the
		// tie by moving the one without CPU instantaneous or memory RSS usage
		// to the beginning.
		return hasMemoryAndCPUInstUsage(&a[j].cinfo)
	}
	return a[i].cinfo.Spec.CreationTime.Before(a[j].cinfo.Spec.CreationTime)
}

// containerID is the identity of a container in a pod.
type containerID struct {
	podRef        statsapi.PodReference
	containerName string
}

// containerInfoWithCgroup contains the ContainerInfo and its cgroup name.
type containerInfoWithCgroup struct {
	cinfo  cadvisorapiv2.ContainerInfo
	cgroup string
}

// hasMemoryAndCPUInstUsage returns true if the specified container info has
// both non-zero CPU instantaneous usage and non-zero memory RSS usage, and
// false otherwise.
func hasMemoryAndCPUInstUsage(info *cadvisorapiv2.ContainerInfo) bool {
	if !info.Spec.HasCpu || !info.Spec.HasMemory {
		return false
	}
	cstat, found := latestContainerStats(info)
	if !found {
		return false
	}
	if cstat.CpuInst == nil {
		return false
	}
	return cstat.CpuInst.Usage.Total != 0 && cstat.Memory.RSS != 0
}

// isPodManagedContainer returns true if the cinfo container is managed by a Pod
func isPodManagedContainer(cinfo *cadvisorapiv2.ContainerInfo) bool {
	podName := kubetypes.GetPodName(cinfo.Spec.Labels)
	podNamespace := kubetypes.GetPodNamespace(cinfo.Spec.Labels)
	managed := podName != "" && podNamespace != ""
	if !managed && podName != podNamespace {
		klog.Warningf(
			"Expect container to have either both podName (%s) and podNamespace (%s) labels, or neither.",
			podName, podNamespace)
	}
	return managed
}

// buildPodRef returns a PodReference that identifies the Pod managing cinfo
func buildPodRef(containerLabels map[string]string) statsapi.PodReference {
	podName := kubetypes.GetPodName(containerLabels)
	podNamespace := kubetypes.GetPodNamespace(containerLabels)
	podUID := kubetypes.GetPodUID(containerLabels)
	return statsapi.PodReference{Name: podName, Namespace: podNamespace, UID: podUID}
}
