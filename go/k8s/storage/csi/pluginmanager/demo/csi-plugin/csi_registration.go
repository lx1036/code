package csi_plugin

import (
	"context"
	"strings"
	"time"

	"k8s.io/klog"
)

const (
	csiTimeout = 2 * time.Minute
)

// PluginHandler is the plugin registration handler interface passed to the
// pluginwatcher module in kubelet
var PluginHandler = &RegistrationHandler{}

// RegistrationHandler is the handler which is fed to the pluginwatcher API.
type RegistrationHandler struct {
}

func (h *RegistrationHandler) ValidatePlugin(pluginName string, endpoint string, versions []string) error {
	klog.Infof("Trying to validate a new CSI Driver with name: %s endpoint: %s versions: %s", pluginName, endpoint, strings.Join(versions, ","))

	return nil
}

// RegisterPlugin is called when a plugin can be registered
func (h *RegistrationHandler) RegisterPlugin(pluginName string, endpoint string, versions []string) error {
	klog.Infof("Register new plugin with name: %s at endpoint: %s", pluginName, endpoint)

	// Get node info from the driver.
	csiClient, err := newCsiDriverClient(csiDriverName(pluginName))
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), csiTimeout)
	defer cancel()

	driverNodeID, maxVolumePerNode, accessibleTopology, err := csiClient.NodeGetInfo(ctx)
	if err != nil {
		if unregErr := unregisterDriver(pluginName); unregErr != nil {
			klog.Error("registrationHandler.RegisterPlugin failed to unregister plugin due to previous error: %v", unregErr)
		}
		return err
	}

	klog.Infof("NodeGetInfo: ")

	return nil
}

// DeRegisterPlugin is called when a plugin removed its socket, signaling
// it is no longer available
func (h *RegistrationHandler) DeRegisterPlugin(pluginName string) {
	klog.Info("registrationHandler.DeRegisterPlugin request for plugin %s", pluginName)
	if err := unregisterDriver(pluginName); err != nil {
		klog.Error("registrationHandler.DeRegisterPlugin failed: %v", err)
	}
}
