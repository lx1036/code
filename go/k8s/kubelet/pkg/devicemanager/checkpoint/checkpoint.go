package checkpoint

import (
	"k8s.io/kubernetes/pkg/kubelet/checkpointmanager"
	"k8s.io/kubernetes/pkg/kubelet/checkpointmanager/checksum"
)

// DeviceManagerCheckpoint defines the operations to retrieve pod devices
type DeviceManagerCheckpoint interface {
	checkpointmanager.Checkpoint
	GetData() ([]PodDevicesEntry, map[string][]string)
}

// PodDevicesEntry connects pod information to devices
type PodDevicesEntry struct {
	PodUID        string
	ContainerName string
	ResourceName  string
	DeviceIDs     []string
	AllocResp     []byte
}

// checkpointData struct is used to store pod to device allocation information
// in a checkpoint file.
// TODO: add version control when we need to change checkpoint format.
type checkpointData struct {
	PodDeviceEntries  []PodDevicesEntry
	RegisteredDevices map[string][]string
}

// Data holds checkpoint data and its checksum
type Data struct {
	Data     checkpointData
	Checksum checksum.Checksum
}

func (d *Data) MarshalCheckpoint() ([]byte, error) {
	panic("implement me")
}

func (d *Data) UnmarshalCheckpoint(blob []byte) error {
	panic("implement me")
}

func (d *Data) VerifyChecksum() error {
	panic("implement me")
}

func (d *Data) GetData() ([]PodDevicesEntry, map[string][]string) {
	panic("implement me")
}

// New returns an instance of Checkpoint
func New(devEntries []PodDevicesEntry,
	devices map[string][]string) DeviceManagerCheckpoint {
	return &Data{
		Data: checkpointData{
			PodDeviceEntries:  devEntries,
			RegisteredDevices: devices,
		},
	}
}
