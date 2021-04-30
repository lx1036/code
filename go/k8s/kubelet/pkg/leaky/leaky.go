// Package leaky holds bits of kubelet that should be internal but have leaked
// out through bad abstractions.  TODO: delete all of this.
package leaky

const (
	// PodInfraContainerName is used in a few places outside of Kubelet, such as indexing
	// into the container info.
	PodInfraContainerName = "POD"
)
