package pluginmanager

import (
	"k8s-lx1036/k8s/storage/csi/pluginmanager/cache"
	"k8s-lx1036/k8s/storage/csi/pluginmanager/pluginwatcher"
	"k8s-lx1036/k8s/storage/csi/pluginmanager/reconciler"
	"k8s.io/client-go/tools/record"

	"k8s.io/kubernetes/pkg/kubelet/config"
)

// PluginManager runs a set of asynchronous loops that figure out which plugins
// need to be registered/deregistered and makes it so.
type PluginManager interface {
	// Starts the plugin manager and all the asynchronous loops that it controls
	Run(sourcesReady config.SourcesReady, stopCh <-chan struct{})

	// AddHandler adds the given plugin handler for a specific plugin type, which
	// will be added to the actual state of world cache so that it can be passed to
	// the desired state of world cache in order to be used during plugin
	// registration/deregistration
	AddHandler(pluginType string, pluginHandler cache.PluginHandler)
}

// pluginManager implements the PluginManager interface
type pluginManager struct {
	// desiredStateOfWorldPopulator (the plugin watcher) runs an asynchronous
	// periodic loop to populate the desiredStateOfWorld.
	desiredStateOfWorldPopulator *pluginwatcher.Watcher

	// reconciler runs an asynchronous periodic loop to reconcile the
	// desiredStateOfWorld with the actualStateOfWorld by triggering register
	// and unregister operations using the operationExecutor.
	reconciler reconciler.Reconciler

	// actualStateOfWorld is a data structure containing the actual state of
	// the world according to the manager: i.e. which plugins are registered.
	// The data structure is populated upon successful completion of register
	// and unregister actions triggered by the reconciler.
	actualStateOfWorld cache.ActualStateOfWorld

	// desiredStateOfWorld is a data structure containing the desired state of
	// the world according to the plugin manager: i.e. what plugins are registered.
	// The data structure is populated by the desired state of the world
	// populator (plugin watcher).
	desiredStateOfWorld cache.DesiredStateOfWorld
}

// NewPluginManager returns a new concrete instance implementing the
// PluginManager interface.
func NewPluginManager(
	sockDir string,
	recorder record.EventRecorder) PluginManager {

	asw := cache.NewActualStateOfWorld()
	dsw := cache.NewDesiredStateOfWorld()

	pm := &pluginManager{
		desiredStateOfWorldPopulator: pluginwatcher.NewWatcher(
			sockDir,
			dsw,
		),
		reconciler:          reconciler,
		desiredStateOfWorld: dsw,
		actualStateOfWorld:  asw,
	}
	return pm
}
