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
