package operationexecutor

import (
	"context"
	"fmt"
	"net"
	"time"

	"k8s-lx1036/k8s/storage/csi/pluginmanager/cache"

	"github.com/pkg/errors"

	"google.golang.org/grpc"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
	registerapi "k8s.io/kubelet/pkg/apis/pluginregistration/v1"
)

const (
	dialTimeoutDuration   = 10 * time.Second
	notifyTimeoutDuration = 5 * time.Second
)

var _ OperationGenerator = &operationGenerator{}

// OperationGenerator interface that extracts out the functions from operation_executor to make it dependency injectable
type OperationGenerator interface {
	// Generates the RegisterPlugin function needed to perform the registration of a plugin
	GenerateRegisterPluginFunc(
		socketPath string,
		timestamp time.Time,
		pluginHandlers map[string]cache.PluginHandler,
		actualStateOfWorldUpdater ActualStateOfWorldUpdater) func() error

	// Generates the UnregisterPlugin function needed to perform the unregistration of a plugin
	GenerateUnregisterPluginFunc(
		pluginInfo cache.PluginInfo,
		actualStateOfWorldUpdater ActualStateOfWorldUpdater) func() error
}

type operationGenerator struct {

	// recorder is used to record events in the API server
	recorder record.EventRecorder
}


func (og operationGenerator) GenerateRegisterPluginFunc(
	socketPath string,
	timestamp time.Time,
	pluginHandlers map[string]cache.PluginHandler,
	actualStateOfWorldUpdater ActualStateOfWorldUpdater) func() error {

	registerPluginFunc := func() error {
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
			return fmt.Errorf("RegisterPlugin error -- failed to get plugin info using RPC GetInfo at socket %s, err: %v", socketPath, err)
		}

		handler, ok := pluginHandlers[infoResp.Type]
		if !ok {
			if err := og.notifyPlugin(client, false, fmt.Sprintf("RegisterPlugin error -- no handler registered for plugin type: %s at socket %s", infoResp.Type, socketPath)); err != nil {
				klog.Errorf("RegisterPlugin error -- failed to send error at socket %s, err: %v", socketPath, err)
				return fmt.Errorf("RegisterPlugin error -- failed to send error at socket %s, err: %v", socketPath, err)
			}
			
			klog.Errorf("RegisterPlugin error -- no handler registered for plugin type: %s at socket %s", infoResp.Type, socketPath)
			return fmt.Errorf("RegisterPlugin error -- no handler registered for plugin type: %s at socket %s", infoResp.Type, socketPath)
		}

		if infoResp.Endpoint == "" {
			infoResp.Endpoint = socketPath
		}
		if err := handler.ValidatePlugin(infoResp.Name, infoResp.Endpoint, infoResp.SupportedVersions); err != nil {
			if err = og.notifyPlugin(client, false, fmt.Sprintf("RegisterPlugin error -- plugin validation failed with err: %v", err)); err != nil {
				klog.Errorf("RegisterPlugin error -- failed to send error at socket %s, err: %v", socketPath, err)
				return fmt.Errorf("RegisterPlugin error -- failed to send error at socket %s, err: %v", socketPath, err)
			}
			
			klog.Errorf("RegisterPlugin error -- pluginHandler.ValidatePluginFunc failed")
			return fmt.Errorf("RegisterPlugin error -- pluginHandler.ValidatePluginFunc failed")
		}

		// We add the plugin to the actual state of world cache before calling a plugin consumer's Register handle
		// so that if we receive a delete event during Register Plugin, we can process it as a DeRegister call.
		err = actualStateOfWorldUpdater.AddPlugin(cache.PluginInfo{
			SocketPath: socketPath,
			Timestamp:  timestamp,
			Handler:    handler,
			Name:       infoResp.Name,
		})
		if err != nil {
			klog.Errorf("RegisterPlugin error -- failed to add plugin at socket %s, err: %v", socketPath, err)
		}
		if err := handler.RegisterPlugin(infoResp.Name, infoResp.Endpoint, infoResp.SupportedVersions); err != nil {
			return og.notifyPlugin(client, false, fmt.Sprintf("RegisterPlugin error -- plugin registration failed with err: %v", err))
		}

		// Notify is called after register to guarantee that even if notify throws an error Register will always be called after validate
		if err := og.notifyPlugin(client, true, ""); err != nil {
			return fmt.Errorf("RegisterPlugin error -- failed to send registration status at socket %s, err: %v", socketPath, err)
		}

		return nil
	}

	return registerPluginFunc
}

func (og operationGenerator) GenerateUnregisterPluginFunc(
	pluginInfo cache.PluginInfo,
	actualStateOfWorldUpdater ActualStateOfWorldUpdater) func() error {
	unregisterPluginFunc := func() error {
		if pluginInfo.Handler == nil {
			return fmt.Errorf("UnregisterPlugin error -- failed to get plugin handler for %s", pluginInfo.SocketPath)
		}

		// We remove the plugin to the actual state of world cache before calling a plugin consumer's Unregister handle
		// so that if we receive a register event during Register Plugin, we can process it as a Register call.
		actualStateOfWorldUpdater.RemovePlugin(pluginInfo.SocketPath)

		pluginInfo.Handler.DeRegisterPlugin(pluginInfo.Name)

		klog.Infof("DeRegisterPlugin called for %s on %v", pluginInfo.Name, pluginInfo.Handler)
		return nil
	}

	return unregisterPluginFunc
}

func (og *operationGenerator) notifyPlugin(client registerapi.RegistrationClient, registered bool, errStr string) error {
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

// NewOperationGenerator is returns instance of operationGenerator
func NewOperationGenerator(recorder record.EventRecorder) OperationGenerator {
	return &operationGenerator{
		recorder: recorder,
	}
}
