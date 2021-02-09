package operationexecutor

import (
	"fmt"
	"errors"
	"net"
	"time"
	"context"
	
	"k8s-lx1036/k8s/storage/csi/pluginmanager/cache"
	
	"google.golang.org/grpc"
	
	"k8s.io/kubernetes/pkg/util/goroutinemap"
	registerapi "k8s.io/kubelet/pkg/apis/pluginregistration/v1"
	"k8s.io/klog/v2"
)

// OperationExecutor defines a set of operations for registering and unregistering
// a plugin that are executed with a NewGoRoutineMap which
// prevents more than one operation from being triggered on the same socket path.
//
// These operations should be idempotent (for example, RegisterPlugin should
// still succeed if the plugin is already registered, etc.). However,
// they depend on the plugin handlers (for each plugin type) to implement this
// behavior.
//
// Once an operation completes successfully, the actualStateOfWorld is updated
// to indicate the plugin is registered/unregistered.
//
// Once the operation is started, since it is executed asynchronously,
// errors are simply logged and the goroutine is terminated without updating
// actualStateOfWorld.
type OperationExecutor interface {
	// RegisterPlugin registers the given plugin using the a handler in the plugin handler map.
	// It then updates the actual state of the world to reflect that.
	RegisterPlugin(socketPath string, timestamp time.Time, pluginHandlers map[string]cache.PluginHandler, actualStateOfWorld ActualStateOfWorldUpdater) error

	// UnregisterPlugin deregisters the given plugin using a handler in the given plugin handler map.
	// It then updates the actual state of the world to reflect that.
	UnregisterPlugin(pluginInfo cache.PluginInfo, actualStateOfWorld ActualStateOfWorldUpdater) error
}

// ActualStateOfWorldUpdater defines a set of operations updating the actual
// state of the world cache after successful registration/deregistration.
type ActualStateOfWorldUpdater interface {
	// AddPlugin add the given plugin in the cache if no existing plugin
	// in the cache has the same socket path.
	// An error will be returned if socketPath is empty.
	AddPlugin(pluginInfo cache.PluginInfo) error

	// RemovePlugin deletes the plugin with the given socket path from the actual
	// state of world.
	// If a plugin does not exist with the given socket path, this is a no-op.
	RemovePlugin(socketPath string)
}

type operationExecutor struct {
	// pendingOperations keeps track of pending attach and detach operations so
	// multiple operations are not started on the same volume
	pendingOperations goroutinemap.GoRoutineMap

	// operationGenerator is an interface that provides implementations for
	// generating volume function
	operationGenerator OperationGenerator
}



// NewOperationExecutor returns a new instance of OperationExecutor.
func NewOperationExecutor(operationGenerator OperationGenerator) OperationExecutor {
	return &operationExecutor{
		pendingOperations:  goroutinemap.NewGoRoutineMap(true /* exponentialBackOffOnError */),
		operationGenerator: operationGenerator,
	}
}
