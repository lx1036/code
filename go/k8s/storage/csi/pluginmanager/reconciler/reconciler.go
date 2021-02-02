package reconciler

import "k8s-lx1036/k8s/storage/csi/pluginmanager/cache"

// Reconciler runs a periodic loop to reconcile the desired state of the world
// with the actual state of the world by triggering register and unregister
// operations.
type Reconciler interface {
	// Starts running the reconciliation loop which executes periodically, checks
	// if plugins that should be registered are register and plugins that should be
	// unregistered are unregistered. If not, it will trigger register/unregister
	// operations to rectify.
	Run(stopCh <-chan struct{})

	// AddHandler adds the given plugin handler for a specific plugin type
	AddHandler(pluginType string, pluginHandler cache.PluginHandler)
}
