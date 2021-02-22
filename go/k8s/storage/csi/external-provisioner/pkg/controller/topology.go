package controller

import (
	"k8s-lx1036/k8s/storage/csi/csi-lib-utils/rpc"
	"k8s-lx1036/k8s/storage/csi/external-provisioner/pkg/features"

	"github.com/container-storage-interface/spec/lib/go/csi"

	utilfeature "k8s.io/apiserver/pkg/util/feature"
)

// SupportsTopology returns whether topology is supported both for plugin and external provisioner
func SupportsTopology(pluginCapabilities rpc.PluginCapabilitySet) bool {
	return pluginCapabilities[csi.PluginCapability_Service_VOLUME_ACCESSIBILITY_CONSTRAINTS] &&
		utilfeature.DefaultFeatureGate.Enabled(features.Topology)
}
