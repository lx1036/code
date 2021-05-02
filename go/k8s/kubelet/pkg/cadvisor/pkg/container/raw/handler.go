package raw

import (
	"k8s-lx1036/k8s/kubelet/pkg/cadvisor/pkg/container"
	"k8s-lx1036/k8s/kubelet/pkg/cadvisor/pkg/container/common"
	"k8s-lx1036/k8s/kubelet/pkg/cadvisor/pkg/container/libcontainer"
	"k8s-lx1036/k8s/kubelet/pkg/cadvisor/pkg/fs"
	v1 "k8s-lx1036/k8s/kubelet/pkg/cadvisor/pkg/info/v1"
)

type rawContainerHandler struct {
	// Name of the container for this handler.
	name               string
	machineInfoFactory v1.MachineInfoFactory

	// Absolute path to the cgroup hierarchies of this container.
	// (e.g.: "cpu" -> "/sys/fs/cgroup/cpu/test")
	cgroupPaths map[string]string

	fsInfo          fs.FsInfo
	externalMounts  []common.Mount
	includedMetrics container.MetricSet

	libcontainerHandler *libcontainer.Handler
}

func (h *rawContainerHandler) ContainerReference() (v1.ContainerReference, error) {
	panic("implement me")
}

func (h *rawContainerHandler) GetSpec() (v1.ContainerSpec, error) {
	panic("implement me")
}

func (h *rawContainerHandler) GetStats() (*v1.ContainerStats, error) {
	panic("implement me")
}

func (h *rawContainerHandler) ListContainers(listType container.ListType) ([]v1.ContainerReference, error) {
	panic("implement me")
}

func (h *rawContainerHandler) ListProcesses(listType container.ListType) ([]int, error) {
	panic("implement me")
}

func (h *rawContainerHandler) GetCgroupPath(resource string) (string, error) {
	panic("implement me")
}

func (h *rawContainerHandler) GetContainerLabels() map[string]string {
	panic("implement me")
}

func (h *rawContainerHandler) GetContainerIPAddress() string {
	panic("implement me")
}

func (h *rawContainerHandler) Exists() bool {
	panic("implement me")
}

func (h *rawContainerHandler) Cleanup() {
	panic("implement me")
}

func (h *rawContainerHandler) Start() {
	panic("implement me")
}

func (h *rawContainerHandler) Type() container.ContainerType {
	panic("implement me")
}

func isRootCgroup(name string) bool {
	return name == "/"
}

func newRawContainerHandler(name string, cgroupSubsystems *libcontainer.CgroupSubsystems,
	machineInfoFactory v1.MachineInfoFactory, fsInfo fs.FsInfo, watcher *common.InotifyWatcher,
	rootFs string, includedMetrics container.MetricSet) (container.ContainerHandler, error) {

	cgroupPaths := common.MakeCgroupPaths(cgroupSubsystems.MountPoints, name)

	cgroupManager, err := libcontainer.NewCgroupManager(name, cgroupPaths)
	if err != nil {
		return nil, err
	}

	var externalMounts []common.Mount
	pid := 0
	if isRootCgroup(name) {
		pid = 1

		// delete pids from cgroup paths because /sys/fs/cgroup/pids/pids.current not exist
		delete(cgroupPaths, "pids")
	}
	handler := libcontainer.NewHandler(cgroupManager, rootFs, pid, includedMetrics)

	return &rawContainerHandler{
		name:                name,
		machineInfoFactory:  machineInfoFactory,
		cgroupPaths:         cgroupPaths,
		fsInfo:              fsInfo,
		externalMounts:      externalMounts,
		includedMetrics:     includedMetrics,
		libcontainerHandler: handler,
	}, nil
}
