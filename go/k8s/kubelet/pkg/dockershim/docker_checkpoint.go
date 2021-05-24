package dockershim

import (
	"encoding/json"

	"k8s-lx1036/k8s/kubelet/pkg/checkpointmanager"
	"k8s-lx1036/k8s/kubelet/pkg/checkpointmanager/checksum"
)

type Protocol string

const (
	// default directory to store pod sandbox checkpoint files
	sandboxCheckpointDir = "sandbox"
	protocolTCP          = Protocol("tcp")
	protocolUDP          = Protocol("udp")
	protocolSCTP         = Protocol("sctp")
	schemaVersion        = "v1"
)

// PortMapping is the port mapping configurations of a sandbox.
type PortMapping struct {
	// Protocol of the port mapping.
	Protocol *Protocol `json:"protocol,omitempty"`
	// Port number within the container.
	ContainerPort *int32 `json:"container_port,omitempty"`
	// Port number on the host.
	HostPort *int32 `json:"host_port,omitempty"`
	// Host ip to expose.
	HostIP string `json:"host_ip,omitempty"`
}

// CheckpointData contains all types of data that can be stored in the checkpoint.
type CheckpointData struct {
	PortMappings []*PortMapping `json:"port_mappings,omitempty"`
	HostNetwork  bool           `json:"host_network,omitempty"`
}

// PodSandboxCheckpoint is the checkpoint structure for a sandbox
type PodSandboxCheckpoint struct {
	// Version of the pod sandbox checkpoint schema.
	Version string `json:"version"`
	// Pod name of the sandbox. Same as the pod name in the Pod ObjectMeta.
	Name string `json:"name"`
	// Pod namespace of the sandbox. Same as the pod namespace in the Pod ObjectMeta.
	Namespace string `json:"namespace"`
	// Data to checkpoint for pod sandbox.
	Data *CheckpointData `json:"data,omitempty"`
	// Checksum is calculated with fnv hash of the checkpoint object with checksum field set to be zero
	Checksum checksum.Checksum `json:"checksum"`
}

func (checkpoint *PodSandboxCheckpoint) MarshalCheckpoint() ([]byte, error) {
	checkpoint.Checksum = checksum.New(*checkpoint.Data)
	return json.Marshal(*checkpoint)
}

func (checkpoint *PodSandboxCheckpoint) UnmarshalCheckpoint(blob []byte) error {
	panic("implement me")
}

func (checkpoint *PodSandboxCheckpoint) VerifyChecksum() error {
	panic("implement me")
}

func (checkpoint *PodSandboxCheckpoint) GetData() (string, string, string, []*PortMapping, bool) {
	panic("implement me")
}

type DockershimCheckpoint interface {
	checkpointmanager.Checkpoint
	GetData() (string, string, string, []*PortMapping, bool)
}

func NewPodSandboxCheckpoint(namespace, name string, data *CheckpointData) DockershimCheckpoint {
	return &PodSandboxCheckpoint{
		Version:   schemaVersion,
		Namespace: namespace,
		Name:      name,
		Data:      data,
	}
}
