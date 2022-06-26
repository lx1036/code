package utils

const (
	// Prefix is the common prefix for all annotations
	Prefix = "io.cilium"

	// V4CIDRName is the annotation name used to store the IPv4
	// pod CIDR in the node's annotations.
	V4CIDRName = Prefix + ".network.ipv4-pod-cidr"
)
