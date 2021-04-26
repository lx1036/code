package container

import (
	"fmt"
	"sync"

	"k8s-lx1036/k8s/kubelet/pkg/cadvisor/pkg/fs"
	"k8s-lx1036/k8s/kubelet/pkg/cadvisor/pkg/info/v1"
	"k8s-lx1036/k8s/kubelet/pkg/cadvisor/pkg/watcher"

	"k8s.io/klog/v2"
)

type Plugin interface {
	// InitializeFSContext is invoked when populating an fs.Context object for a new manager.
	// A returned error here is fatal.
	InitializeFSContext(context *fs.Context) error

	// Register is invoked when starting a manager. It can optionally return a container watcher.
	// A returned error is logged, but is not fatal.
	Register(factory v1.MachineInfoFactory, fsInfo fs.FsInfo, includedMetrics MetricSet) (watcher.ContainerWatcher, error)
}

type ContainerHandlerFactory interface {
	// Create a new ContainerHandler using this factory. CanHandleAndAccept() must have returned true.
	NewContainerHandler(name string, inHostNamespace bool) (c ContainerHandler, err error)

	// Returns whether this factory can handle and accept the specified container.
	CanHandleAndAccept(name string) (handle bool, accept bool, err error)

	// Name of the factory.
	String() string

	// Returns debugging information. Map of lines per category.
	DebugInfo() map[string][]string
}

var (
	factories     = map[watcher.ContainerWatchSource][]ContainerHandlerFactory{}
	factoriesLock sync.RWMutex
)

// All registered auth provider plugins.
var pluginsLock sync.Mutex
var plugins = make(map[string]Plugin)

// Register a ContainerHandlerFactory. These should be registered from least general to most general
// as they will be asked in order whether they can handle a particular container.
func RegisterContainerHandlerFactory(factory ContainerHandlerFactory, watchTypes []watcher.ContainerWatchSource) {
	factoriesLock.Lock()
	defer factoriesLock.Unlock()

	for _, watchType := range watchTypes {
		factories[watchType] = append(factories[watchType], factory)
	}
}

func RegisterPlugin(name string, plugin Plugin) error {
	pluginsLock.Lock()
	defer pluginsLock.Unlock()
	if _, found := plugins[name]; found {
		return fmt.Errorf("Plugin %q was registered twice", name)
	}
	klog.V(4).Infof("Registered Plugin %q", name)
	plugins[name] = plugin
	return nil
}

func InitializePlugins(factory v1.MachineInfoFactory, fsInfo fs.FsInfo, includedMetrics MetricSet) []watcher.ContainerWatcher {
	pluginsLock.Lock()
	defer pluginsLock.Unlock()

	var containerWatchers []watcher.ContainerWatcher
	for name, plugin := range plugins {
		watcher, err := plugin.Register(factory, fsInfo, includedMetrics)
		if err != nil {
			klog.V(5).Infof("Registration of the %s container factory failed: %v", name, err)
		}
		if watcher != nil {
			containerWatchers = append(containerWatchers, watcher)
		}
	}
	return containerWatchers
}

// Returns whether there are any container handler factories registered.
func HasFactories() bool {
	factoriesLock.Lock()
	defer factoriesLock.Unlock()

	return len(factories) != 0
}

// Create a new ContainerHandler for the specified container.
func NewContainerHandler(name string, watchType watcher.ContainerWatchSource, inHostNamespace bool) (ContainerHandler, bool, error) {
	factoriesLock.RLock()
	defer factoriesLock.RUnlock()

	// Create the ContainerHandler with the first factory that supports it.
	for _, factory := range factories[watchType] {
		canHandle, canAccept, err := factory.CanHandleAndAccept(name)
		if err != nil {
			klog.V(4).Infof("Error trying to work out if we can handle %s: %v", name, err)
		}
		if canHandle {
			if !canAccept {
				klog.V(3).Infof("Factory %q can handle container %q, but ignoring.", factory, name)
				return nil, false, nil
			}
			klog.V(3).Infof("Using factory %q for container %q", factory, name)
			handle, err := factory.NewContainerHandler(name, inHostNamespace)
			return handle, canAccept, err
		}
		klog.V(4).Infof("Factory %q was unable to handle container %q", factory, name)
	}

	return nil, false, fmt.Errorf("no known factory can handle creation of container")
}
