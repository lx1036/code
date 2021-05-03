package raw

import (
	"flag"
	"fmt"
	"strings"

	"k8s-lx1036/k8s/kubelet/pkg/cadvisor/pkg/container"
	"k8s-lx1036/k8s/kubelet/pkg/cadvisor/pkg/container/common"
	"k8s-lx1036/k8s/kubelet/pkg/cadvisor/pkg/container/libcontainer"
	"k8s-lx1036/k8s/kubelet/pkg/cadvisor/pkg/fs"
	v1 "k8s-lx1036/k8s/kubelet/pkg/cadvisor/pkg/info/v1"
	watch "k8s-lx1036/k8s/kubelet/pkg/cadvisor/pkg/watcher"

	"k8s.io/klog/v2"
)

var dockerOnly = flag.Bool("docker_only", false, "Only report docker containers in addition to root stats")

type rawFactory struct {
	// Factory for machine information.
	machineInfoFactory v1.MachineInfoFactory

	// Information about the cgroup subsystems.
	cgroupSubsystems *libcontainer.CgroupSubsystems

	// Information about mounted filesystems.
	fsInfo fs.FsInfo

	// Watcher for inotify events.
	watcher *common.InotifyWatcher

	// List of metrics to be included.
	includedMetrics map[container.MetricKind]struct{}

	// List of raw container cgroup path prefix whitelist.
	rawPrefixWhiteList []string
}

func (f *rawFactory) NewContainerHandler(name string, inHostNamespace bool) (c container.ContainerHandler, err error) {
	rootFs := "/"
	if !inHostNamespace {
		rootFs = "/rootfs"
	}

	return newRawContainerHandler(name, f.cgroupSubsystems, f.machineInfoFactory, f.fsInfo, f.watcher, rootFs, f.includedMetrics)
}

func (f *rawFactory) CanHandleAndAccept(name string) (handle bool, accept bool, err error) {
	if name == "/" {
		return true, true, nil
	}

	if *dockerOnly && f.rawPrefixWhiteList[0] == "" {
		return true, false, nil
	}
	for _, prefix := range f.rawPrefixWhiteList {
		if strings.HasPrefix(name, prefix) {
			return true, true, nil
		}
	}

	return true, false, nil
}

func (f *rawFactory) String() string {
	return "raw"
}

func (f *rawFactory) DebugInfo() map[string][]string {
	panic("implement me")
}

func Register(machineInfoFactory v1.MachineInfoFactory, fsInfo fs.FsInfo,
	includedMetrics map[container.MetricKind]struct{}, rawPrefixWhiteList []string) error {
	cgroupSubsystems, err := libcontainer.GetCgroupSubsystems(includedMetrics)
	if err != nil {
		return fmt.Errorf("failed to get cgroup subsystems: %v", err)
	}
	if len(cgroupSubsystems.Mounts) == 0 {
		return fmt.Errorf("failed to find supported cgroup mounts for the raw factory")
	}

	watcher, err := common.NewInotifyWatcher()
	if err != nil {
		return err
	}

	klog.V(1).Infof("Registering Raw factory")
	factory := &rawFactory{
		machineInfoFactory: machineInfoFactory,
		fsInfo:             fsInfo,
		cgroupSubsystems:   &cgroupSubsystems,
		watcher:            watcher,
		includedMetrics:    includedMetrics,
		rawPrefixWhiteList: rawPrefixWhiteList,
	}
	container.RegisterContainerHandlerFactory(factory, []watch.ContainerWatchSource{watch.Raw})

	return nil
}
