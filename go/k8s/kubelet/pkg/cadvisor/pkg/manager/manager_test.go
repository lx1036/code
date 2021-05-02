package manager

import (
	"strings"
	"testing"
	"time"

	"k8s-lx1036/k8s/kubelet/pkg/cadvisor/pkg/cache/memory"
	"k8s-lx1036/k8s/kubelet/pkg/cadvisor/pkg/container"
	"k8s-lx1036/k8s/kubelet/pkg/cadvisor/pkg/container/docker"
	containertest "k8s-lx1036/k8s/kubelet/pkg/cadvisor/pkg/container/testing"
	"k8s-lx1036/k8s/kubelet/pkg/cadvisor/pkg/info/v1"
	itest "k8s-lx1036/k8s/kubelet/pkg/cadvisor/pkg/info/v1/test"
	"k8s-lx1036/k8s/kubelet/pkg/cadvisor/pkg/info/v2"
	"k8s-lx1036/k8s/kubelet/pkg/cadvisor/pkg/utils/sysfs/fakesysfs"

	"github.com/stretchr/testify/assert"
	clock "k8s.io/utils/clock/testing"
)

// Expect a manager with the specified containers and query. Returns the manager, map of ContainerInfo objects,
// and map of MockContainerHandler objects.}
func expectManagerWithSubContainers(containers []string, query *v1.ContainerInfoRequest,
	t *testing.T) (*manager, map[string]*v1.ContainerInfo, map[string]*containertest.MockContainerHandler) {
	infosMap := make(map[string]*v1.ContainerInfo, len(containers))
	handlerMap := make(map[string]*containertest.MockContainerHandler, len(containers))

	for _, c := range containers {
		infosMap[c] = itest.GenerateRandomContainerInfo(c, 4, query, 1*time.Second)
	}

	memoryCache := memory.New(time.Duration(query.NumStats)*time.Second, nil)
	sysfs := &fakesysfs.FakeSysFs{}
	m := createManagerAndAddSubContainers(memoryCache, sysfs, containers,
		func(h *containertest.MockContainerHandler) {
			cinfo := infosMap[h.Name]
			ref, err := h.ContainerReference()
			if err != nil {
				t.Error(err)
			}

			cInfo := v1.ContainerInfo{
				ContainerReference: ref,
			}
			for _, stat := range cinfo.Stats {
				err = memoryCache.AddStats(&cInfo, stat)
				if err != nil {
					t.Error(err)
				}
			}
			spec := cinfo.Spec
			h.On("GetSpec").Return(spec, nil).Once()
			handlerMap[h.Name] = h
		},
		t,
	)

	return m, infosMap, handlerMap
}

func createManagerAndAddSubContainers(memoryCache *memory.InMemoryCache, sysfs *fakesysfs.FakeSysFs,
	containers []string, f func(*containertest.MockContainerHandler), t *testing.T) *manager {
	container.ClearContainerHandlerFactories()
	mif := &manager{
		containers:   make(map[namespacedContainerName]*containerData),
		quitChannels: make([]chan error, 0, 2),
		memoryCache:  memoryCache,
	}

	subcontainers1 := []v1.ContainerReference{
		{Name: "/kubepods/besteffort"},
		{Name: "/kubepods/burstable"},
	}
	subcontainers2 := []v1.ContainerReference(nil)
	subcontainers3 := []v1.ContainerReference{
		{Name: "/kubepods/burstable/pod01042b28-179d-446a-954a-7266557e12cd"},
		{Name: "/kubepods/burstable/pod01042b28-179d-446a-954a-7266557e12ce"},
	}
	subcontainers4 := []v1.ContainerReference{
		{Name: "/kubepods/burstable/pod01042b28-179d-446a-954a-7266557e12cd/22f44d2a517778590e2d8bcafafe501f79e8a509e5b6de70b7700c4d37722bce"},
		{Name: "/kubepods/burstable/pod01042b28-179d-446a-954a-7266557e12cd/ae9465f98d275998e148b6fc12f5f92e5d4a64fca0d255f6dc3a13cc6f93a10f"},
	}

	subcontainers5 := []v1.ContainerReference(nil)
	subcontainers6 := []v1.ContainerReference(nil)

	subcontainerList := [][]v1.ContainerReference{subcontainers1, subcontainers2, subcontainers3, subcontainers4, subcontainers5, subcontainers6}

	for idx, name := range containers {
		mockHandler := containertest.NewMockContainerHandler(name)
		spec := itest.GenerateRandomContainerSpec(4)
		mockHandler.On("GetSpec").Return(spec, nil).Once()
		mockHandler.On("ListContainers", container.ListSelf).Return(
			subcontainerList[idx],
			nil,
		)
		cont, err := newContainerData(name, memoryCache, mockHandler, false, 60*time.Second,
			true, clock.NewFakeClock(time.Now()))
		if err != nil {
			t.Fatal(err)
		}
		mif.containers[namespacedContainerName{Name: name}] = cont
		// Add Docker containers under their namespace.
		if strings.HasPrefix(name, "/docker") {
			mif.containers[namespacedContainerName{
				Namespace: docker.DockerNamespace,
				Name:      strings.TrimPrefix(name, "/docker/"),
			}] = cont
		}
		f(mockHandler)
	}
	return mif
}

func TestSubContainersInfoError(t *testing.T) {
	containers := []string{
		"/kubepods",
		"/kubepods/besteffort",
		"/kubepods/burstable",
		"/kubepods/burstable/pod01042b28-179d-446a-954a-7266557e12cd",
		"/kubepods/burstable/pod01042b28-179d-446a-954a-7266557e12cd/22f44d2a517778590e2d8bcafafe501f79e8a509e5b6de70b7700c4d37722bce",
		"/kubepods/burstable/pod01042b28-179d-446a-954a-7266557e12cd/ae9465f98d275998e148b6fc12f5f92e5d4a64fca0d255f6dc3a13cc6f93a10f",
	}

	query := &v1.ContainerInfoRequest{
		NumStats: 1,
	}

	m, _, _ := expectManagerWithSubContainers(containers, query, t)
	result, err := m.SubcontainersInfo("/kubepods", query)
	if err != nil {
		t.Fatalf("expected to succeed: %s", err)
	}

	if len(result) != len(containers) {
		t.Errorf("expected to received containers: %v, but received: %v", containers, result)
	}

	totalBurstable := 0
	burstableCount := 0
	totalBesteffort := 0
	besteffortCount := 0

	for _, res := range result {
		found := false
		if res.Name == "/kubepods/burstable" {
			totalBurstable = len(res.Subcontainers)
		} else if res.Name == "/kubepods/besteffort" {
			totalBesteffort = len(res.Subcontainers)
		} else if strings.HasPrefix(res.Name, "/kubepods/burstable") && len(res.Name) == len("/kubepods/burstable/pod01042b28-179d-446a-954a-7266557e12cd") {
			burstableCount++
		} else if strings.HasPrefix(res.Name, "/kubepods/besteffort") && len(res.Name) == len("/kubepods/besteffort/pod01042b28-179d-446a-954a-7266557e12cd") {
			besteffortCount++
		}
		for _, name := range containers {
			if res.Name == name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("unexpected container %q in result, expected one of %v", res.Name, containers)
		}
	}

	assert.NotEqual(t, totalBurstable, burstableCount)
	assert.Equal(t, totalBesteffort, besteffortCount)
}

// Expect a manager with the specified containers and query. Returns the manager, map of ContainerInfo objects,
// and map of MockContainerHandler objects.}
func expectManagerWithContainersV2(containers []string, query *v1.ContainerInfoRequest, t *testing.T) (*manager,
	map[string]*v1.ContainerInfo, map[string]*containertest.MockContainerHandler) {
	infosMap := make(map[string]*v1.ContainerInfo, len(containers))
	handlerMap := make(map[string]*containertest.MockContainerHandler, len(containers))

	for _, containerName := range containers {
		infosMap[containerName] = itest.GenerateRandomContainerInfo(containerName, 4, query, 1*time.Second)
	}

	memoryCache := memory.New(time.Duration(query.NumStats)*time.Second, nil)
	sysfs := &fakesysfs.FakeSysFs{}
	m := createManagerAndAddContainers(
		memoryCache,
		sysfs,
		containers,
		func(h *containertest.MockContainerHandler) {
			cinfo := infosMap[h.Name]
			ref, err := h.ContainerReference()
			if err != nil {
				t.Error(err)
			}

			cInfo := v1.ContainerInfo{
				ContainerReference: ref,
			}

			for _, stat := range cinfo.Stats {
				err = memoryCache.AddStats(&cInfo, stat)
				if err != nil {
					t.Error(err)
				}
			}
			spec := cinfo.Spec

			h.On("GetSpec").Return(
				spec,
				nil,
			).Once()
			handlerMap[h.Name] = h
		},
		t,
	)

	return m, infosMap, handlerMap
}

func createManagerAndAddContainers(
	memoryCache *memory.InMemoryCache,
	sysfs *fakesysfs.FakeSysFs,
	containers []string,
	f func(*containertest.MockContainerHandler),
	t *testing.T,
) *manager {
	container.ClearContainerHandlerFactories()
	m := &manager{
		containers:   make(map[namespacedContainerName]*containerData),
		quitChannels: make([]chan error, 0, 2),
		memoryCache:  memoryCache,
	}
	for _, name := range containers {
		mockHandler := containertest.NewMockContainerHandler(name)
		spec := itest.GenerateRandomContainerSpec(4)
		mockHandler.On("GetSpec").Return(
			spec,
			nil,
		).Once()
		cont, err := newContainerData(name, memoryCache, mockHandler, false, 60*time.Second,
			true, clock.NewFakeClock(time.Now()))
		if err != nil {
			t.Fatal(err)
		}
		m.containers[namespacedContainerName{
			Name: name,
		}] = cont
		// Add Docker containers under their namespace.
		if strings.HasPrefix(name, "/docker") {
			m.containers[namespacedContainerName{
				Namespace: docker.DockerNamespace,
				Name:      strings.TrimPrefix(name, "/docker/"),
			}] = cont
		}
		f(mockHandler)
	}
	return m
}

func TestGetContainerInfoV2(t *testing.T) {
	containers := []string{
		"/",
		"/c1",
		"/c2",
	}

	options := v2.RequestOptions{
		IdType:    v2.TypeName,
		Count:     1,
		Recursive: true,
	}
	query := &v1.ContainerInfoRequest{
		NumStats: 2,
	}

	m, _, handlerMap := expectManagerWithContainersV2(containers, query, t)

	infos, err := m.GetContainerInfoV2("/", options)
	if err != nil {
		t.Fatalf("GetContainerInfoV2 failed: %v", err)
	}

	for container, handler := range handlerMap {
		handler.AssertExpectations(t)
		info, ok := infos[container]
		assert.True(t, ok, "Missing info for container %q", container)
		assert.NotEqual(t, v2.ContainerSpec{}, info.Spec, "Empty spec for container %q", container)
		assert.NotEmpty(t, info.Stats, "Missing stats for container %q", container)
	}
}
