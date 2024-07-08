package identity

import (
	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/labels"
)

const (
	NodeLocalIdentityType    = "node_local"
	ReservedIdentityType     = "reserved"
	ClusterLocalIdentityType = "cluster_local"
	WellKnownIdentityType    = "well_known"
)

// Identity is the representation of the security context for a particular set of
// labels.
type Identity struct {
	// Identity's ID.
	ID NumericIdentity `json:"id"`
	// Set of labels that belong to this Identity.
	Labels labels.Labels `json:"labels"`

	// LabelArray contains the same labels as Labels in a form of a list, used
	// for faster lookup.
	LabelArray labels.LabelArray `json:"-"`

	// CIDRLabel is the primary identity label when the identity represents
	// a CIDR. The Labels field will consist of all matching prefixes, e.g.
	// 10.0.0.0/8
	// 10.0.0.0/7
	// 10.0.0.0/6
	// [...]
	// reserved:world
	//
	// The CIDRLabel field will only contain 10.0.0.0/8
	CIDRLabel labels.Labels `json:"-"`

	// ReferenceCount counts the number of references pointing to this
	// identity. This field is used by the owning cache of the identity.
	ReferenceCount int `json:"-"`
}

// NewIdentity creates a new identity
func NewIdentity(id NumericIdentity, lbls labels.Labels) *Identity {
	var lblArray labels.LabelArray

	if lbls != nil {
		lblArray = lbls.LabelArray()
	}
	return &Identity{ID: id, Labels: lbls, LabelArray: lblArray}
}
