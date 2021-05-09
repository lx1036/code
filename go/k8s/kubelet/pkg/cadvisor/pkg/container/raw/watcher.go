package raw

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"k8s-lx1036/k8s/kubelet/pkg/cadvisor/pkg/container/common"
	"k8s-lx1036/k8s/kubelet/pkg/cadvisor/pkg/container/libcontainer"
	"k8s-lx1036/k8s/kubelet/pkg/cadvisor/pkg/watcher"

	"github.com/fsnotify/fsnotify"
	"k8s.io/klog/v2"
)

// INFO: 这里可以使用 "github.com/fsnotify/fsnotify"(支持macos) 替换 common.InotifyWatcher, 直接参考 go/k8s/storage/csi/pluginmanager/pluginwatcher/plugin_watcher.go

type rawContainerWatcher struct {
	// Absolute path to the root of the cgroup hierarchies
	cgroupPaths map[string]string

	cgroupSubsystems *libcontainer.CgroupSubsystems

	// Inotify event watcher.
	watcher *common.InotifyWatcher

	// Signal for watcher thread to stop.
	stopWatcher chan error
}

// INFO: watch dir and all subdirectories
func (w *rawContainerWatcher) watchDirectory(events chan watcher.ContainerEvent, dir string, containerName string) (bool, error) {
	// Don't watch .mount cgroups because they never have containers as sub-cgroups.  A single container
	// can have many .mount cgroups associated with it which can quickly exhaust the inotify watches on a node.
	if strings.HasSuffix(containerName, ".mount") {
		return false, nil
	}
	alreadyWatching, err := w.watcher.AddWatch(containerName, dir)
	if err != nil {
		return alreadyWatching, err
	}

	// Remove the watch if further operations failed.
	cleanup := true
	defer func() {
		if cleanup {
			_, err := w.watcher.RemoveWatch(containerName, dir)
			if err != nil {
				klog.Warningf("Failed to remove inotify watch for %q: %v", dir, err)
			}
		}
	}()

	// Watch subdirectories as well.
	entries, err := ioutil.ReadDir(dir)
	if err != nil {
		return alreadyWatching, err
	}
	for _, entry := range entries {
		if entry.IsDir() {
			entryPath := path.Join(dir, entry.Name())
			subcontainerName := path.Join(containerName, entry.Name())
			alreadyWatchingSubDir, err := w.watchDirectory(events, entryPath, subcontainerName)
			if err != nil {
				klog.Errorf("Failed to watch directory %q: %v", entryPath, err)
				if os.IsNotExist(err) {
					// The directory may have been removed before watching. Try to watch the other
					// subdirectories. (https://github.com/kubernetes/kubernetes/issues/28997)
					continue
				}
				return alreadyWatching, err
			}

			// since we already missed the creation event for this directory, publish an event here.
			if !alreadyWatchingSubDir {
				go func() {
					events <- watcher.ContainerEvent{
						EventType:   watcher.ContainerAdd,
						Name:        subcontainerName,
						WatchSource: watcher.Raw,
					}
				}()
			}
		}
	}

	cleanup = false
	return alreadyWatching, nil
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

func (w *rawContainerWatcher) processEvent(event fsnotify.Event, events chan watcher.ContainerEvent) error {
	// Convert the inotify event type to a container create or delete.
	var eventType watcher.ContainerEventType
	switch {
	case (event.Op & fsnotify.Create) > 0:
		eventType = watcher.ContainerAdd
	case (event.Op & fsnotify.Remove) > 0:
		eventType = watcher.ContainerDelete
	case (event.Op & fsnotify.Rename) > 0:
		eventType = watcher.ContainerDelete
	default:
		// Ignore other events.
		return nil
	}

	// Derive the container name from the path name.
	var containerName string
	for _, mount := range w.cgroupSubsystems.Mounts {
		mountLocation := path.Clean(mount.Mountpoint) + "/"
		if strings.HasPrefix(event.Name, mountLocation) {
			containerName = event.Name[len(mountLocation)-1:]
			break
		}
	}
	if containerName == "" {
		return fmt.Errorf("unable to detect container from watch event on directory %q", event.Name)
	}

	// Maintain the watch for the new or deleted container.
	switch eventType {
	case watcher.ContainerAdd:
		// New container was created, watch it.
		alreadyWatched, err := w.watchDirectory(events, event.Name, containerName)
		if err != nil {
			return err
		}

		// Only report container creation once.
		if alreadyWatched {
			return nil
		}
	case watcher.ContainerDelete:
		// Container was deleted, stop watching for it.
		lastWatched, err := w.watcher.RemoveWatch(containerName, event.Name)
		if err != nil {
			return err
		}

		// Only report container deletion once.
		if !lastWatched {
			return nil
		}
	default:
		return fmt.Errorf("unknown event type %v", eventType)
	}

	// Deliver the event.
	events <- watcher.ContainerEvent{
		EventType:   eventType,
		Name:        containerName,
		WatchSource: watcher.Raw,
	}

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
