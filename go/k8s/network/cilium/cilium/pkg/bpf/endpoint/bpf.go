package endpoint

import (
	"k8s-lx1036/k8s/network/cilium/cilium/pkg/bpf"
	"k8s-lx1036/k8s/network/cilium/cilium/pkg/bpf/maps/policymap"
)

// policyMapPath returns the path to the policy map of endpoint.
func (e *Endpoint) policyMapPath() string {
	return bpf.LocalMapPath(policymap.MapName, e.ID)
}

// InitPolicyMap creates the policy map in the kernel.
func (e *Endpoint) InitPolicyMap() error {
	_, err := policymap.Create(e.policyMapPath())
	return err
}
