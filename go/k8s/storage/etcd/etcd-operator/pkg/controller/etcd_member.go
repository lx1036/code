package controller

type Member struct {
	Name string
	// Kubernetes namespace this member runs in.
	Namespace string
	// ID field can be 0, which is unknown ID.
	// We know the ID of a member when we get the member information from etcd,
	// but not from Kubernetes pod list.
	ID uint64

	SecurePeer   bool
	SecureClient bool

	// ClusterDomain is the DNS name of the cluster. E.g. .cluster.local.
	ClusterDomain string
}

type MemberSet map[string]*Member
