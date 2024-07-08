package identity

import (
	"math"

	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/labels"
)

// NumericIdentity is the numeric representation of a security identity.
//
// Bits:
//
//	 0-15: identity identifier
//	16-23: cluster identifier
//	   24: LocalIdentityFlag: Indicates that the identity has a local scope
type NumericIdentity uint32

// MaxNumericIdentity is the maximum value of a NumericIdentity.
const MaxNumericIdentity = math.MaxUint32

const (
	// IdentityUnknown represents an unknown identity
	IdentityUnknown NumericIdentity = iota

	// ReservedIdentityHost represents the local host
	ReservedIdentityHost

	// ReservedIdentityWorld represents any endpoint outside of the cluster
	ReservedIdentityWorld

	// ReservedIdentityUnmanaged represents unmanaged endpoints.
	ReservedIdentityUnmanaged

	// ReservedIdentityHealth represents the local cilium-health endpoint
	ReservedIdentityHealth

	// ReservedIdentityInit is the identity given to endpoints that have not
	// received any labels yet.
	ReservedIdentityInit

	// ReservedIdentityRemoteNode is the identity given to all nodes in
	// local and remote clusters except for the local node.
	ReservedIdentityRemoteNode

	// ReservedIdentityKubeAPIServer is the identity given to remote node(s) which
	// have backend(s) serving the kube-apiserver running.
	ReservedIdentityKubeAPIServer

	// ReservedIdentityIngress is the identity given to the IP used as the source
	// address for connections from Ingress proxies.
	ReservedIdentityIngress
)

var (
	reservedIdentityLabels = map[NumericIdentity]labels.Labels{
		ReservedIdentityHost:       labels.LabelHost,
		ReservedIdentityWorld:      labels.LabelWorld,
		ReservedIdentityUnmanaged:  labels.NewLabelsFromModel([]string{"reserved:" + labels.IDNameUnmanaged}),
		ReservedIdentityHealth:     labels.LabelHealth,
		ReservedIdentityInit:       labels.NewLabelsFromModel([]string{"reserved:" + labels.IDNameInit}),
		ReservedIdentityRemoteNode: labels.LabelRemoteNode,
		ReservedIdentityKubeAPIServer: labels.Map2Labels(map[string]string{
			labels.LabelKubeAPIServer.String(): "",
			labels.LabelRemoteNode.String():    "",
		}, ""),
		ReservedIdentityIngress: labels.LabelIngress,
	}
)

// iterateReservedIdentityLabels iterates over all reservedIdentityLabels and
// executes the given function for each key, value pair in
// reservedIdentityLabels.
func iterateReservedIdentityLabels(f func(_ NumericIdentity, _ labels.Labels)) {
	for ni, lbls := range reservedIdentityLabels {
		f(ni, lbls)
	}
}
