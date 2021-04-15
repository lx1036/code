package devicemanager

import (
	"k8s-lx1036/k8s/kubelet/pkg/devicemanager/checkpoint"

	"k8s.io/apimachinery/pkg/util/sets"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

type deviceAllocateInfo struct {
	// deviceIds contains device Ids allocated to this container for the given resourceName.
	deviceIds sets.String
	// allocResp contains cached rpc AllocateResponse.
	allocResp *pluginapi.ContainerAllocateResponse
}

type resourceAllocateInfo map[string]deviceAllocateInfo // Keyed by resourceName.
type containerDevices map[string]resourceAllocateInfo   // Keyed by containerName.
type podDevices map[string]containerDevices             // Keyed by podUID.

// Populates podDevices from the passed in checkpointData.
func (pdev podDevices) fromCheckpointData(data []checkpoint.PodDevicesEntry) {

}

// Returns all of devices allocated to the pods being tracked, keyed by resourceName.
func (pdev podDevices) devices() map[string]sets.String {
	ret := make(map[string]sets.String)

	return ret
}
