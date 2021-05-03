package stats

import (
	"testing"

	statsapi "k8s-lx1036/k8s/kubelet/pkg/apis/stats/v1alpha1"
	"k8s-lx1036/k8s/kubelet/pkg/cadvisor"
	cadvisorapiv2 "k8s-lx1036/k8s/kubelet/pkg/cadvisor/pkg/info/v2"
	cadvisortest "k8s-lx1036/k8s/kubelet/pkg/cadvisor/testing"
	kubecontainer "k8s-lx1036/k8s/kubelet/pkg/container"
	containertest "k8s-lx1036/k8s/kubelet/pkg/container/testing"
	"k8s-lx1036/k8s/kubelet/pkg/leaky"
	serverstats "k8s-lx1036/k8s/kubelet/pkg/server/stats"
	statustest "k8s-lx1036/k8s/kubelet/pkg/status/testing"
	kubelettypes "k8s-lx1036/k8s/kubelet/pkg/types"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func TestCadvisorListPodCPUAndMemoryStats(t *testing.T) {
	const (
		seedRoot              = 0
		seedRuntime           = 100
		seedKubelet           = 200
		seedMisc              = 300
		seedPod0Infra         = 1000
		seedPod0Container0    = 2000
		seedPod0Container1    = 2001
		seedPod1Infra         = 3000
		seedPod1Container     = 4000
		seedPod2Infra         = 5000
		seedPod2Container     = 6000
		seedEphemeralVolume1  = 10000
		seedEphemeralVolume2  = 10001
		seedPersistentVolume1 = 20000
		seedPersistentVolume2 = 20001
	)

	// cadvisor
	runtime := kubelettypes.DockerContainerRuntime
	runtimeEndpoint := "unix:///var/run/docker.sock"
	imageFsInfoProvider := cadvisor.NewImageFsInfoProvider(runtime, runtimeEndpoint)
	rootPath := ""
	cgroupRoots := []string{"/kubepods"}
	cadvisorClient, err := cadvisor.New(imageFsInfoProvider, rootPath, cgroupRoots, true)
	if err != nil {
		panic(err)
	}

	// resourceAnalyzer volume
	ephemeralVolumes := []statsapi.VolumeStats{getPodVolumeStats(seedEphemeralVolume1, "ephemeralVolume1"),
		getPodVolumeStats(seedEphemeralVolume2, "ephemeralVolume2")}
	persistentVolumes := []statsapi.VolumeStats{getPodVolumeStats(seedPersistentVolume1, "persistentVolume1"),
		getPodVolumeStats(seedPersistentVolume2, "persistentVolume2")}
	volumeStats := serverstats.PodVolumeStats{
		EphemeralVolumes:  ephemeralVolumes,
		PersistentVolumes: persistentVolumes,
	}
	resourceAnalyzer := &fakeResourceAnalyzer{podVolumeStats: volumeStats}

	p := NewCadvisorStatsProvider(cadvisorClient, resourceAnalyzer, nil, nil, nil, nil)
	pods, err := p.ListPodCPUAndMemoryStats()
	assert.NoError(t, err)
	assert.Equal(t, 3, len(pods))

	// TODO: 继续完善代码测试

}

func TestCadvisorListPodStats(test *testing.T) {
	const (
		seedRoot              = 0
		seedRuntime           = 100
		seedKubelet           = 200
		seedMisc              = 300
		seedPod0Infra         = 1000
		seedPod0Container0    = 2000
		seedPod0Container1    = 2001
		seedPod1Infra         = 3000
		seedPod1Container     = 4000
		seedPod2Infra         = 5000
		seedPod2Container     = 6000
		seedPod3Infra         = 7000
		seedPod3Container0    = 8000
		seedPod3Container1    = 8001
		seedEphemeralVolume1  = 10000
		seedEphemeralVolume2  = 10001
		seedPersistentVolume1 = 20000
		seedPersistentVolume2 = 20001
	)
	const (
		namespace0 = "test0"
		namespace2 = "test2"
	)
	const (
		pName0 = "pod0"
		pName1 = "pod1"
		pName2 = "pod0" // ensure pName2 conflicts with pName0, but is in a different namespace
		pName3 = "pod3"
	)
	const (
		cName00 = "c0"
		cName01 = "c1"
		cName10 = "c0" // ensure cName10 conflicts with cName02, but is in a different pod
		cName20 = "c1" // ensure cName20 conflicts with cName01, but is in a different pod + namespace
		cName30 = "c0-init"
		cName31 = "c1"
	)
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

	runtime := kubelettypes.DockerContainerRuntime
	runtimeEndpoint := "unix:///var/run/docker.sock"
	imageFsInfoProvider := cadvisor.NewImageFsInfoProvider(runtime, runtimeEndpoint)
	rootPath := ""
	cgroupRoots := []string{"/kubepods"}
	cadvisorClient, err := cadvisor.New(imageFsInfoProvider, rootPath, cgroupRoots, true)
	if err != nil {
		panic(err)
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

	// volume
	ephemeralVolumes := []statsapi.VolumeStats{getPodVolumeStats(seedEphemeralVolume1, "ephemeralVolume1"),
		getPodVolumeStats(seedEphemeralVolume2, "ephemeralVolume2")}
	persistentVolumes := []statsapi.VolumeStats{getPodVolumeStats(seedPersistentVolume1, "persistentVolume1"),
		getPodVolumeStats(seedPersistentVolume2, "persistentVolume2")}
	volumeStats := serverstats.PodVolumeStats{
		EphemeralVolumes:  ephemeralVolumes,
		PersistentVolumes: persistentVolumes,
	}
	resourceAnalyzer := &fakeResourceAnalyzer{podVolumeStats: volumeStats}

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

	prf0 := statsapi.PodReference{Name: pName0, Namespace: namespace0, UID: "UID" + pName0}
	prf1 := statsapi.PodReference{Name: pName1, Namespace: namespace0, UID: "UID" + pName1}
	prf2 := statsapi.PodReference{Name: pName2, Namespace: namespace2, UID: "UID" + pName2}
	prf3 := statsapi.PodReference{Name: pName3, Namespace: namespace0, UID: "UID" + pName3}
	indexPods := make(map[statsapi.PodReference]statsapi.PodStats, len(pods))
	for _, pod := range pods {
		indexPods[pod.PodRef] = pod
	}
	// Validate Pod0 Results
	ps, found := indexPods[prf0]
	assert.True(test, found)
	assert.Len(test, ps.Containers, 2)
	indexCon := make(map[string]statsapi.ContainerStats, len(ps.Containers))
	for _, con := range ps.Containers {
		indexCon[con.Name] = con
	}
	con := indexCon[cName00]
	assert.EqualValues(test, testTime(creationTime, seedPod0Container0).Unix(), con.StartTime.Time.Unix())
	checkCPUStats(test, "Pod0Container0", seedPod0Container0, con.CPU)
	checkMemoryStats(test, "Pod0Conainer0", seedPod0Container0, containerInfos["/pod0-c0"], con.Memory)
	con = indexCon[cName01]
	assert.EqualValues(test, testTime(creationTime, seedPod0Container1).Unix(), con.StartTime.Time.Unix())
	checkCPUStats(test, "Pod0Container1", seedPod0Container1, con.CPU)
	checkMemoryStats(test, "Pod0Container1", seedPod0Container1, containerInfos["/pod0-c1"], con.Memory)
	assert.EqualValues(test, p0Time.Unix(), ps.StartTime.Time.Unix())
	checkNetworkStats(test, "Pod0", seedPod0Infra, ps.Network)
	checkEphemeralStats(test, "Pod0", []int{seedPod0Container0, seedPod0Container1}, []int{seedEphemeralVolume1, seedEphemeralVolume2}, ps.EphemeralStorage)
	if ps.CPU != nil {
		checkCPUStats(test, "Pod0", seedPod0Infra, ps.CPU)
	}
	if ps.Memory != nil {
		checkMemoryStats(test, "Pod0", seedPod0Infra, containerInfos["/pod0-i"], ps.Memory)
	}

	// Validate Pod1 Results
	ps, found = indexPods[prf1]
	assert.True(test, found)
	assert.Len(test, ps.Containers, 1)
	con = ps.Containers[0]
	assert.Equal(test, cName10, con.Name)
	checkCPUStats(test, "Pod1Container0", seedPod1Container, con.CPU)
	checkMemoryStats(test, "Pod1Container0", seedPod1Container, containerInfos["/pod1-c0"], con.Memory)
	checkNetworkStats(test, "Pod1", seedPod1Infra, ps.Network)

	// Validate Pod2 Results
	ps, found = indexPods[prf2]
	assert.True(test, found)
	assert.Len(test, ps.Containers, 1)
	con = ps.Containers[0]
	assert.Equal(test, cName20, con.Name)
	checkCPUStats(test, "Pod2Container0", seedPod2Container, con.CPU)
	checkMemoryStats(test, "Pod2Container0", seedPod2Container, containerInfos["/pod2-c0"], con.Memory)
	checkNetworkStats(test, "Pod2", seedPod2Infra, ps.Network)

	// Validate Pod3 Results
	ps, found = indexPods[prf3]
	assert.True(test, found)
	assert.Len(test, ps.Containers, 2)
	indexCon = make(map[string]statsapi.ContainerStats, len(ps.Containers))
	for _, con := range ps.Containers {
		indexCon[con.Name] = con
	}
	con = indexCon[cName31]
	assert.Equal(test, cName31, con.Name)
	checkCPUStats(test, "Pod3Container1", seedPod3Container1, con.CPU)
	checkMemoryStats(test, "Pod3Container1", seedPod3Container1, containerInfos["/pod3-c1"], con.Memory)
	con = indexCon[cName30]
	assert.Equal(test, cName30, con.Name)
	checkEmptyCPUStats(test, "Pod3Container0", seedPod3Container0, con.CPU)
	checkEmptyMemoryStats(test, "Pod3Container0", seedPod3Container0, containerInfos["/pod3-c0-init"], con.Memory)
}
