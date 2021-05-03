package raw

import (
	"fmt"

	"k8s-lx1036/k8s/kubelet/pkg/cadvisor/pkg/container/common"
	"k8s-lx1036/k8s/kubelet/pkg/cadvisor/pkg/container/libcontainer"
	"k8s-lx1036/k8s/kubelet/pkg/cadvisor/pkg/watcher"

	"k8s.io/klog/v2"
)

type rawContainerWatcher struct {
	// Absolute path to the root of the cgroup hierarchies
	cgroupPaths map[string]string

	cgroupSubsystems *libcontainer.CgroupSubsystems

	// Inotify event watcher.
	watcher *common.InotifyWatcher

	// Signal for watcher thread to stop.
	stopWatcher chan error
}

func (w *rawContainerWatcher) Start(events chan watcher.ContainerEvent) error {
	// Watch this container (all its cgroups) and all subdirectories.
	watched := make([]string, 0)
	for _, cgroupPath := range w.cgroupPaths {
		_, err := w.watchDirectory(events, cgroupPath, "/")
		if err != nil {
			for _, watchedCgroupPath := range watched {
				_, removeErr := w.watcher.RemoveWatch("/", watchedCgroupPath)
				if removeErr != nil {
					klog.Warningf("Failed to remove inotify watch for %q with error: %v", watchedCgroupPath, removeErr)
				}
			}
			return err
		}
		watched = append(watched, cgroupPath)
	}

	// Process the events received from the kernel.
	go func() {
		for {
			select {
			case event := <-w.watcher.Event():
				err := w.processEvent(event, events)
				if err != nil {
					klog.Warningf("Error while processing event (%+v): %v", event, err)
				}
			case err := <-w.watcher.Error():
				klog.Warningf("Error while watching %q: %v", "/", err)
			case <-w.stopWatcher:
				err := w.watcher.Close()
				if err == nil {
					w.stopWatcher <- err
					return
				}
			}
		}
	}()

	return nil
}

func (w *rawContainerWatcher) Stop() error {
	panic("implement me")
}

func NewRawContainerWatcher() (watcher.ContainerWatcher, error) {
	cgroupSubsystems, err := libcontainer.GetAllCgroupSubsystems()
	if err != nil {
		return nil, fmt.Errorf("failed to get cgroup subsystems: %v", err)
	}
	if len(cgroupSubsystems.Mounts) == 0 {
		return nil, fmt.Errorf("failed to find supported cgroup mounts for the raw factory")
	}

	inotifyWatcher, err := common.NewInotifyWatcher()
	if err != nil {
		return nil, err
	}

	rawWatcher := &rawContainerWatcher{
		cgroupPaths:      common.MakeCgroupPaths(cgroupSubsystems.MountPoints, "/"),
		cgroupSubsystems: &cgroupSubsystems,
		watcher:          inotifyWatcher,
		stopWatcher:      make(chan error),
	}

	return rawWatcher, nil
}
