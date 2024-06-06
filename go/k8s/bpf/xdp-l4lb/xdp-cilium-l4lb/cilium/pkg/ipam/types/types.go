package types

// AllocationIP is an IP which is available for allocation, or already
// has been allocated
type AllocationIP struct {
    // Owner is the owner of the IP. This field is set if the IP has been
    // allocated. It will be set to the pod name or another identifier
    // representing the usage of the IP
    //
    // The owner field is left blank for an entry in Spec.IPAM.Pool and
    // filled out as the IP is used and also added to Status.IPAM.Used.
    //
    // +optional
    Owner string `json:"owner,omitempty"`

    // Resource is set for both available and allocated IPs, it represents
    // what resource the IP is associated with, e.g. in combination with
    // AWS ENI, this will refer to the ID of the ENI
    //
    // +optional
    Resource string `json:"resource,omitempty"`
}

// AllocationMap is a map of allocated IPs indexed by IP
type AllocationMap map[string]AllocationIP

// IPAMSpec is the IPAM specification of the node
//
// This structure is embedded into v2.CiliumNode
type IPAMSpec struct {
    // Pool is the list of IPs available to the node for allocation. When
    // an IP is used, the IP will remain on this list but will be added to
    // Status.IPAM.Used
    //
    // +optional
    Pool AllocationMap `json:"pool,omitempty"`

    // PodCIDRs is the list of CIDRs available to the node for allocation.
    // When an IP is used, the IP will be added to Status.IPAM.Used
    //
    // +optional
    PodCIDRs []string `json:"podCIDRs,omitempty"`

    // MinAllocate is the minimum number of IPs that must be allocated when
    // the node is first bootstrapped. It defines the minimum base socket
    // of addresses that must be available. After reaching this watermark,
    // the PreAllocate and MaxAboveWatermark logic takes over to continue
    // allocating IPs.
    //
    // +kubebuilder:validation:Minimum=0
    MinAllocate int `json:"min-allocate,omitempty"`

    // MaxAllocate is the maximum number of IPs that can be allocated to the
    // node. When the current amount of allocated IPs will approach this value,
    // the considered value for PreAllocate will decrease down to 0 in order to
    // not attempt to allocate more addresses than defined.
    //
    // +kubebuilder:validation:Minimum=0
    MaxAllocate int `json:"max-allocate,omitempty"`

    // PreAllocate defines the number of IP addresses that must be
    // available for allocation in the IPAMspec. It defines the buffer of
    // addresses available immediately without requiring cilium-operator to
    // get involved.
    //
    // +kubebuilder:validation:Minimum=0
    PreAllocate int `json:"pre-allocate,omitempty"`

    // MaxAboveWatermark is the maximum number of addresses to allocate
    // beyond the addresses needed to reach the PreAllocate watermark.
    // Going above the watermark can help reduce the number of API calls to
    // allocate IPs, e.g. when a new ENI is allocated, as many secondary
    // IPs as possible are allocated. Limiting the amount can help reduce
    // waste of IPs.
    //
    // +kubebuilder:validation:Minimum=0
    MaxAboveWatermark int `json:"max-above-watermark,omitempty"`

    // PodCIDRAllocationThreshold defines the minimum number of free IPs which
    // must be available to this node via its pod CIDR pool. If the total number
    // of IP addresses in the pod CIDR pool is less than this value, the pod
    // CIDRs currently in-use by this node will be marked as depleted and
    // cilium-operator will allocate a new pod CIDR to this node.
    // This value effectively defines the buffer of IP addresses available
    // immediately without requiring cilium-operator to get involved.
    //
    // +kubebuilder:validation:Minimum=0
    PodCIDRAllocationThreshold int `json:"pod-cidr-allocation-threshold,omitempty"`

    // PodCIDRReleaseThreshold defines the maximum number of free IPs which may
    // be available to this node via its pod CIDR pool. While the total number
    // of free IP addresses in the pod CIDR pool is larger than this value,
    // cilium-agent will attempt to release currently unused pod CIDRs.
    //
    // +kubebuilder:validation:Minimum=0
    PodCIDRReleaseThreshold int `json:"pod-cidr-release-threshold,omitempty"`
}
