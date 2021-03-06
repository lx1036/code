package cache

import (
	"fmt"
	"sync"
	"time"

	"k8s.io/klog/v2"
)

// ActualStateOfWorld defines a set of thread-safe operations for the kubelet
// plugin manager's actual state of the world cache.
// This cache contains a map of socket file path to plugin information of
// all plugins attached to this node.
type ActualStateOfWorld interface {
	// GetRegisteredPlugins generates and returns a list of plugins
	// that are successfully registered plugins in the current actual state of world.
	GetRegisteredPlugins() []PluginInfo

	// AddPlugin add the given plugin in the cache.
	// An error will be returned if socketPath of the PluginInfo object is empty.
	// Note that this is different from desired world cache's AddOrUpdatePlugin
	// because for the actual state of world cache, there won't be a scenario where
	// we need to update an existing plugin if the timestamps don't match. This is
	// because the plugin should have been unregistered in the reconciller and therefore
	// removed from the actual state of world cache first before adding it back into
	// the actual state of world cache again with the new timestamp
	AddPlugin(pluginInfo PluginInfo) error

	// RemovePlugin deletes the plugin with the given socket path from the actual
	// state of world.
	// If a plugin does not exist with the given socket path, this is a no-op.
	RemovePlugin(socketPath string)

	// PluginExists checks if the given plugin exists in the current actual
	// state of world cache with the correct timestamp
	PluginExistsWithCorrectTimestamp(pluginInfo PluginInfo) bool
}

// PluginInfo holds information of a plugin
type PluginInfo struct {
	SocketPath string
	Timestamp  time.Time
	Handler    PluginHandler
	Name       string
}

type actualStateOfWorld struct {
	// socketFileToInfo is a map containing the set of successfully registered plugins
	// The keys are plugin socket file paths. The values are PluginInfo objects
	socketFileToInfo map[string]PluginInfo
	sync.RWMutex
}

var _ ActualStateOfWorld = &actualStateOfWorld{}

// NewActualStateOfWorld returns a new instance of ActualStateOfWorld
func NewActualStateOfWorld() ActualStateOfWorld {
	return &actualStateOfWorld{
		socketFileToInfo: make(map[string]PluginInfo),
	}
}

func (asw *actualStateOfWorld) AddPlugin(pluginInfo PluginInfo) error {
	asw.Lock()
	defer asw.Unlock()

	klog.Infof("add plugin SocketPath %s at time %s into actualStateOfWorld map", pluginInfo.SocketPath, pluginInfo.Timestamp.String())

	if pluginInfo.SocketPath == "" {
		return fmt.Errorf("socket path is empty")
	}
	if _, ok := asw.socketFileToInfo[pluginInfo.SocketPath]; ok {
		klog.Infof("Plugin (Path %s) exists in actual state cache", pluginInfo.SocketPath)
	}
	asw.socketFileToInfo[pluginInfo.SocketPath] = pluginInfo
	return nil
}

func (asw *actualStateOfWorld) RemovePlugin(socketPath string) {
	asw.Lock()
	defer asw.Unlock()

	delete(asw.socketFileToInfo, socketPath)
}

func (asw *actualStateOfWorld) GetRegisteredPlugins() []PluginInfo {
	asw.RLock()
	defer asw.RUnlock()

	var currentPlugins []PluginInfo
	for _, pluginInfo := range asw.socketFileToInfo {
		currentPlugins = append(currentPlugins, pluginInfo)
	}
	return currentPlugins
}

func (asw *actualStateOfWorld) PluginExistsWithCorrectTimestamp(pluginInfo PluginInfo) bool {
	asw.RLock()
	defer asw.RUnlock()

	// We need to check both if the socket file path exists, and the timestamp
	// matches the given plugin (from the desired state cache) timestamp
	actualStatePlugin, exists := asw.socketFileToInfo[pluginInfo.SocketPath]
	return exists && (actualStatePlugin.Timestamp == pluginInfo.Timestamp)
}
