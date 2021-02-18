package reconciler

import (
	"context"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"k8s-lx1036/k8s/storage/csi/pluginmanager/cache"

	"github.com/pkg/errors"
	"google.golang.org/grpc"

	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
	registerapi "k8s.io/kubelet/pkg/apis/pluginregistration/v1"
)

const (
	dialTimeoutDuration   = 10 * time.Second
	notifyTimeoutDuration = 5 * time.Second
)

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

type reconciler struct {
	loopSleepDuration   time.Duration
	desiredStateOfWorld cache.DesiredStateOfWorld
	actualStateOfWorld  cache.ActualStateOfWorld
	handlers            map[string]cache.PluginHandler
	sync.RWMutex
}

func (rc *reconciler) getHandlers() map[string]cache.PluginHandler {
	rc.RLock()
	defer rc.RUnlock()

	return rc.handlers
}

func (rc *reconciler) logDesiredActualStateOfWorld() {
	var desiredSocketPath []string
	for _, pluginToRegister := range rc.desiredStateOfWorld.GetPluginsToRegister() {
		desiredSocketPath = append(desiredSocketPath, pluginToRegister.SocketPath)
	}

	klog.Infof("desiredSocketPaths in current reconcile: %s", strings.Join(desiredSocketPath, ","))

	var actualSocketPath []string
	for _, registeredPlugin := range rc.actualStateOfWorld.GetRegisteredPlugins() {
		actualSocketPath = append(actualSocketPath, registeredPlugin.SocketPath)
	}

	klog.Infof("actualSocketPaths in current reconcile: %s", strings.Join(actualSocketPath, ","))
}

func (rc *reconciler) reconcile() {
	// Unregisterations are triggered before registrations
	klog.Infof("reconcile in every time %s", rc.loopSleepDuration.String())

	rc.logDesiredActualStateOfWorld()

	// Ensure plugins that should be unregistered are unregistered.
	for _, registeredPlugin := range rc.actualStateOfWorld.GetRegisteredPlugins() {
		unregisterPlugin := false
		if !rc.desiredStateOfWorld.PluginExists(registeredPlugin.SocketPath) {
			unregisterPlugin = true
		} else {
			// We also need to unregister the plugins that exist in both actual state of world
			// and desired state of world cache, but the timestamps don't match.
			// Iterate through desired state of world plugins and see if there's any plugin
			// with the same socket path but different timestamp.
			for _, dswPlugin := range rc.desiredStateOfWorld.GetPluginsToRegister() {
				if dswPlugin.SocketPath == registeredPlugin.SocketPath && dswPlugin.Timestamp != registeredPlugin.Timestamp {
					klog.Infof("An updated version of plugin has been found: plugin %s created at %s", dswPlugin.SocketPath, dswPlugin.Timestamp.String())
					unregisterPlugin = true
					break
				}
			}
		}

		if unregisterPlugin {
			err := rc.UnregisterPlugin(registeredPlugin, rc.actualStateOfWorld)
			if err != nil {
				klog.Errorf("failed to unregister plugin: %v", err)
				continue
			}

			klog.Infof("unregister plugin %s successfully", registeredPlugin.SocketPath)
		}
	}

	// Ensure plugins that should be registered are registered
	for _, pluginToRegister := range rc.desiredStateOfWorld.GetPluginsToRegister() {
		if !rc.actualStateOfWorld.PluginExistsWithCorrectTimestamp(pluginToRegister) {
			err := rc.RegisterPlugin(pluginToRegister.SocketPath, pluginToRegister.Timestamp, rc.getHandlers(), rc.actualStateOfWorld)
			if err != nil {
				klog.Errorf("failed to register plugin with %v", err)
				continue
			}

			klog.Infof("register plugin %s created at %s successfully", pluginToRegister.SocketPath, pluginToRegister.Timestamp.String())
		}
	}
}

// dial RegistrationClient.GetInfo()
func (rc *reconciler) RegisterPlugin(socketPath string, timestamp time.Time, pluginHandlers map[string]cache.PluginHandler, actualStateOfWorld cache.ActualStateOfWorld) error {
	client, conn, err := dial(socketPath, dialTimeoutDuration)
	if err != nil {
		klog.Errorf("RegisterPlugin error -- dial failed at socket %s, err: %v", socketPath, err)
		return fmt.Errorf("RegisterPlugin error -- dial failed at socket %s, err: %v", socketPath, err)
	}
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	infoResp, err := client.GetInfo(ctx, &registerapi.InfoRequest{})
	if err != nil {
		klog.Errorf("RegisterPlugin error -- failed to get plugin info using RPC GetInfo at socket %s, err: %v", socketPath, err)
		return err
	}

	handler, ok := pluginHandlers[infoResp.Type]
	if !ok {
		if err = rc.notifyPlugin(client, false, fmt.Sprintf("RegisterPlugin error -- no handler registered for plugin type: %s at socket %s", infoResp.Type, socketPath)); err != nil {
			klog.Errorf("RegisterPlugin error -- failed to send error at socket %s, err: %v", socketPath, err)
			return err
		}

		klog.Errorf("RegisterPlugin error -- no handler registered for plugin type: %s at socket %s", infoResp.Type, socketPath)
		return err
	}

	if infoResp.Endpoint == "" {
		infoResp.Endpoint = socketPath
	}
	if err = handler.ValidatePlugin(infoResp.Name, infoResp.Endpoint, infoResp.SupportedVersions); err != nil {
		if err = rc.notifyPlugin(client, false, fmt.Sprintf("RegisterPlugin error -- plugin validation failed with err: %v", err)); err != nil {
			klog.Errorf("RegisterPlugin error -- failed to send error at socket %s, err: %v", socketPath, err)
			return err
		}

		klog.Errorf("RegisterPlugin error -- pluginHandler.ValidatePluginFunc failed")
		return err
	}

	// We add the plugin to the actual state of world cache before calling a plugin consumer's Register handle
	// so that if we receive a delete event during Register Plugin, we can process it as a DeRegister call.
	err = actualStateOfWorld.AddPlugin(cache.PluginInfo{
		SocketPath: socketPath,
		Timestamp:  timestamp,
		Handler:    handler,
		Name:       infoResp.Name,
	})
	if err != nil {
		klog.Errorf("RegisterPlugin error -- failed to add plugin at socket %s, err: %v", socketPath, err)
	}
	if err = handler.RegisterPlugin(infoResp.Name, infoResp.Endpoint, infoResp.SupportedVersions); err != nil {
		return rc.notifyPlugin(client, false, fmt.Sprintf("RegisterPlugin error -- plugin registration failed with err: %v", err))
	}

	// Notify is called after register to guarantee that even if notify throws an error Register will always be called after validate
	if err := rc.notifyPlugin(client, true, ""); err != nil {
		return fmt.Errorf("RegisterPlugin error -- failed to send registration status at socket %s, err: %v", socketPath, err)
	}

	return nil
}

func (rc *reconciler) UnregisterPlugin(pluginInfo cache.PluginInfo, actualStateOfWorld cache.ActualStateOfWorld) error {
	if pluginInfo.Handler == nil {
		return fmt.Errorf("UnregisterPlugin error -- failed to get plugin handler for %s", pluginInfo.SocketPath)
	}

	// We remove the plugin to the actual state of world cache before calling a plugin consumer's Unregister handle
	// so that if we receive a register event during Register Plugin, we can process it as a Register call.
	actualStateOfWorld.RemovePlugin(pluginInfo.SocketPath)
	pluginInfo.Handler.DeRegisterPlugin(pluginInfo.Name)

	klog.Infof("DeRegisterPlugin called for %s on %v", pluginInfo.Name, pluginInfo.Handler)
	return nil
}

func (rc *reconciler) notifyPlugin(client registerapi.RegistrationClient, registered bool, errStr string) error {
	ctx, cancel := context.WithTimeout(context.Background(), notifyTimeoutDuration)
	defer cancel()

	status := &registerapi.RegistrationStatus{
		PluginRegistered: registered,
		Error:            errStr,
	}

	if _, err := client.NotifyRegistrationStatus(ctx, status); err != nil {
		return errors.Wrap(err, errStr)
	}

	if errStr != "" {
		return errors.New(errStr)
	}

	return nil
}

// Dial establishes the gRPC communication with the picked up plugin socket. https://godoc.org/google.golang.org/grpc#Dial
func dial(unixSocketPath string, timeout time.Duration) (registerapi.RegistrationClient, *grpc.ClientConn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	c, err := grpc.DialContext(ctx, unixSocketPath, grpc.WithInsecure(), grpc.WithBlock(),
		grpc.WithContextDialer(func(ctx context.Context, addr string) (net.Conn, error) {
			return (&net.Dialer{}).DialContext(ctx, "unix", addr)
		}),
	)

	if err != nil {
		return nil, nil, fmt.Errorf("failed to dial socket %s, err: %v", unixSocketPath, err)
	}

	return registerapi.NewRegistrationClient(c), c, nil
}

func (rc *reconciler) Run(stopCh <-chan struct{}) {
	wait.Until(func() {
		rc.reconcile()
	}, rc.loopSleepDuration, stopCh)
}

func (rc *reconciler) AddHandler(pluginType string, pluginHandler cache.PluginHandler) {
	rc.Lock()
	defer rc.Unlock()

	rc.handlers[pluginType] = pluginHandler
}

// NewReconciler returns a new instance of Reconciler.
//
// loopSleepDuration - the amount of time the reconciler loop sleeps between
//   successive executions
//   syncDuration - the amount of time the syncStates sleeps between
//   successive executions
// operationExecutor - used to trigger register/unregister operations safely
//   (prevents more than one operation from being triggered on the same
//   socket path)
// desiredStateOfWorld - cache containing the desired state of the world
// actualStateOfWorld - cache containing the actual state of the world
func NewReconciler(
	loopSleepDuration time.Duration,
	desiredStateOfWorld cache.DesiredStateOfWorld,
	actualStateOfWorld cache.ActualStateOfWorld) Reconciler {
	return &reconciler{
		loopSleepDuration:   loopSleepDuration,
		desiredStateOfWorld: desiredStateOfWorld,
		actualStateOfWorld:  actualStateOfWorld,
		handlers:            make(map[string]cache.PluginHandler),
	}
}
