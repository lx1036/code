package stats

import (
	"testing"

	cadvisorapiv2 "k8s-lx1036/k8s/kubelet/pkg/cadvisor/pkg/info/v2"
	cadvisortest "k8s-lx1036/k8s/kubelet/pkg/cadvisor/testing"
	kubecontainer "k8s-lx1036/k8s/kubelet/pkg/container"
	containertest "k8s-lx1036/k8s/kubelet/pkg/container/testing"
	"k8s-lx1036/k8s/kubelet/pkg/leaky"
	statustest "k8s-lx1036/k8s/kubelet/pkg/status/testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func TestCadvisorListPodStats(test *testing.T) {

	containerInfos := map[string]cadvisorapiv2.ContainerInfo{
		"/":              getTestContainerInfo(seedRoot, "", "", ""),
		"/docker-daemon": getTestContainerInfo(seedRuntime, "", "", ""),
		"/kubelet":       getTestContainerInfo(seedKubelet, "", "", ""),
		"/system":        getTestContainerInfo(seedMisc, "", "", ""),
		// Pod0 - Namespace0
		"/pod0-i":  getTestContainerInfo(seedPod0Infra, pName0, namespace0, leaky.PodInfraContainerName),
		"/pod0-c0": getTestContainerInfo(seedPod0Container0, pName0, namespace0, cName00),
		"/pod0-c1": getTestContainerInfo(seedPod0Container1, pName0, namespace0, cName01),
		// Pod1 - Namespace0
		"/pod1-i":  getTestContainerInfo(seedPod1Infra, pName1, namespace0, leaky.PodInfraContainerName),
		"/pod1-c0": getTestContainerInfo(seedPod1Container, pName1, namespace0, cName10),
		// Pod2 - Namespace2
		"/pod2-i":                        getTestContainerInfo(seedPod2Infra, pName2, namespace2, leaky.PodInfraContainerName),
		"/pod2-c0":                       getTestContainerInfo(seedPod2Container, pName2, namespace2, cName20),
		"/kubepods/burstable/podUIDpod0": getTestContainerInfo(seedPod0Infra, pName0, namespace0, leaky.PodInfraContainerName),
		"/kubepods/podUIDpod1":           getTestContainerInfo(seedPod1Infra, pName1, namespace0, leaky.PodInfraContainerName),
		// Pod3 - Namespace0
		"/pod3-i":       getTestContainerInfo(seedPod3Infra, pName3, namespace0, leaky.PodInfraContainerName),
		"/pod3-c0-init": getTestContainerInfo(seedPod3Container0, pName3, namespace0, cName30),
		"/pod3-c1":      getTestContainerInfo(seedPod3Container1, pName3, namespace0, cName31),
	}

	// rootfs
	const (
		rootfsCapacity    = uint64(10000000)
		rootfsAvailable   = uint64(5000000)
		rootfsInodesFree  = uint64(1000)
		rootfsInodes      = uint64(2000)
		imagefsCapacity   = uint64(20000000)
		imagefsAvailable  = uint64(8000000)
		imagefsInodesFree = uint64(2000)
		imagefsInodes     = uint64(4000)
	)
	freeRootfsInodes := rootfsInodesFree
	totalRootfsInodes := rootfsInodes
	rootfs := cadvisorapiv2.FsInfo{
		Capacity:   rootfsCapacity,
		Available:  rootfsAvailable,
		InodesFree: &freeRootfsInodes,
		Inodes:     &totalRootfsInodes,
	}
	// imagefs
	freeImagefsInodes := imagefsInodesFree
	totalImagefsInodes := imagefsInodes
	imagefs := cadvisorapiv2.FsInfo{
		Capacity:   imagefsCapacity,
		Available:  imagefsAvailable,
		InodesFree: &freeImagefsInodes,
		Inodes:     &totalImagefsInodes,
	}

	// INFO: mock cadvisor, 可以借鉴
	options := cadvisorapiv2.RequestOptions{
		IdType:    cadvisorapiv2.TypeName,
		Count:     2,
		Recursive: true,
	}
	mockCadvisor := new(cadvisortest.Mock)
	mockCadvisor.
		On("ContainerInfoV2", "/", options).Return(containerInfos, nil).
		On("RootFsInfo").Return(rootfs, nil).
		On("ImagesFsInfo").Return(imagefs, nil)

	// INFO: mock runtime, 也可以借鉴
	mockRuntime := new(containertest.Mock)
	mockRuntime.
		On("ImageStats").Return(&kubecontainer.ImageStats{TotalStorageBytes: 123}, nil)

	resourceAnalyzer := &fakeResourceAnalyzer{podVolumeStats: volumeStats}

	const (
		pName0 = "pod0"
		pName1 = "pod1"
		pName2 = "pod0" // ensure pName2 conflicts with pName0, but is in a different namespace
		pName3 = "pod3"
	)
	p0Time := metav1.Now()
	p1Time := metav1.Now()
	p2Time := metav1.Now()
	p3Time := metav1.Now()
	mockStatus := new(statustest.MockStatusProvider)
	mockStatus.On("GetPodStatus", types.UID("UID"+pName0)).Return(v1.PodStatus{StartTime: &p0Time}, true)
	mockStatus.On("GetPodStatus", types.UID("UID"+pName1)).Return(v1.PodStatus{StartTime: &p1Time}, true)
	mockStatus.On("GetPodStatus", types.UID("UID"+pName2)).Return(v1.PodStatus{StartTime: &p2Time}, true)
	mockStatus.On("GetPodStatus", types.UID("UID"+pName3)).Return(v1.PodStatus{StartTime: &p3Time}, true)

	p := NewCadvisorStatsProvider(mockCadvisor, resourceAnalyzer, nil, nil, mockRuntime, mockStatus)
	pods, err := p.ListPodStats()
	assert.NoError(test, err)
	assert.Equal(test, 4, len(pods))

}
