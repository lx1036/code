package cache

import (
	"fmt"
	"sync"
	"time"

	"k8s.io/klog"
)

// DesiredStateOfWorld defines a set of thread-safe operations for the kubelet
// plugin manager's desired state of the world cache.
// This cache contains a map of socket file path to plugin information of
// all plugins attached to this node.
type DesiredStateOfWorld interface {
	// AddOrUpdatePlugin add the given plugin in the cache if it doesn't already exist.
	// If it does exist in the cache, then the timestamp of the PluginInfo object in the cache will be updated.
	// An error will be returned if socketPath is empty.
	AddOrUpdatePlugin(socketPath string) error

	// RemovePlugin deletes the plugin with the given socket path from the desired
	// state of world.
	// If a plugin does not exist with the given socket path, this is a no-op.
	RemovePlugin(socketPath string)

	// GetPluginsToRegister generates and returns a list of plugins
	// in the current desired state of world.
	GetPluginsToRegister() []PluginInfo

	// PluginExists checks if the given socket path exists in the current desired
	// state of world cache
	PluginExists(socketPath string) bool
}

type desiredStateOfWorld struct {

	// socketFileToInfo is a map containing the set of successfully registered plugins
	// The keys are plugin socket file paths. The values are PluginInfo objects
	socketFileToInfo map[string]PluginInfo
	sync.RWMutex
}

var _ DesiredStateOfWorld = &desiredStateOfWorld{}

// NewDesiredStateOfWorld returns a new instance of DesiredStateOfWorld.
func NewDesiredStateOfWorld() DesiredStateOfWorld {
	return &desiredStateOfWorld{
		socketFileToInfo: make(map[string]PluginInfo),
	}
}

func (dsw *desiredStateOfWorld) RemovePlugin(socketPath string) {
	dsw.Lock()
	defer dsw.Unlock()

	delete(dsw.socketFileToInfo, socketPath)
}

func (dsw *desiredStateOfWorld) GetPluginsToRegister() []PluginInfo {
	dsw.RLock()
	defer dsw.RUnlock()

	var pluginsToRegister []PluginInfo
	for _, pluginInfo := range dsw.socketFileToInfo {
		pluginsToRegister = append(pluginsToRegister, pluginInfo)
	}
	return pluginsToRegister
}

func (dsw *desiredStateOfWorld) PluginExists(socketPath string) bool {
	dsw.RLock()
	defer dsw.RUnlock()

	_, exists := dsw.socketFileToInfo[socketPath]
	return exists
}

func (dsw *desiredStateOfWorld) AddOrUpdatePlugin(socketPath string) error {
	dsw.Lock()
	defer dsw.Unlock()

	if socketPath == "" {
		return fmt.Errorf("socket path is empty")
	}
	if _, ok := dsw.socketFileToInfo[socketPath]; ok {
		klog.Infof("Plugin (Path %s) exists in actual state cache, timestamp will be updated", socketPath)
	}

	// Update the PluginInfo object.
	// Note that we only update the timestamp in the desired state of world, not the actual state of world
	// because in the reconciler, we need to check if the plugin in the actual state of world is the same
	// version as the plugin in the desired state of world
	dsw.socketFileToInfo[socketPath] = PluginInfo{
		SocketPath: socketPath,
		Timestamp:  time.Now(),
	}
	return nil
}

// Generate a detailed error msg for logs
func generatePluginMsgDetailed(prefixMsg, suffixMsg, socketPath, details string) (detailedMsg string) {
	return fmt.Sprintf("%v for plugin at %q %v %v", prefixMsg, socketPath, details, suffixMsg)
}

// Generate a simplified error msg for events and a detailed error msg for logs
func generatePluginMsg(prefixMsg, suffixMsg, socketPath, details string) (simpleMsg, detailedMsg string) {
	simpleMsg = fmt.Sprintf("%v for plugin at %q %v", prefixMsg, socketPath, suffixMsg)
	return simpleMsg, generatePluginMsgDetailed(prefixMsg, suffixMsg, socketPath, details)
}

// GenerateMsgDetailed returns detailed msgs for plugins to register
// that can be used in logs.
// The msg format follows the pattern "<prefixMsg> <plugin details> <suffixMsg>"
func (plugin *PluginInfo) GenerateMsgDetailed(prefixMsg, suffixMsg string) (detailedMsg string) {
	detailedStr := fmt.Sprintf("(plugin details: %v)", plugin)
	return generatePluginMsgDetailed(prefixMsg, suffixMsg, plugin.SocketPath, detailedStr)
}

// GenerateErrorDetailed returns detailed errors for plugins to register
// that can be used in logs.
// The msg format follows the pattern "<prefixMsg> <plugin details>: <err> ",
func (plugin *PluginInfo) GenerateErrorDetailed(prefixMsg string, err error) (detailedErr error) {
	return fmt.Errorf(plugin.GenerateMsgDetailed(prefixMsg, errSuffix(err)))
}

// Generates an error string with the format ": <err>" if err exists
func errSuffix(err error) string {
	errStr := ""
	if err != nil {
		errStr = fmt.Sprintf(": %v", err)
	}
	return errStr
}
