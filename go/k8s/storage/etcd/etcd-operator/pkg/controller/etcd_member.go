package controller

import "fmt"

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

func (member *Member) peerScheme() string {
	if member.SecurePeer {
		return "https"
	}

	return "http"
}

func (member *Member) Addr() string {
	return fmt.Sprintf("%s", member.Name)
}

func (member *Member) PeerURL() string {
	return fmt.Sprintf("%s://%s:2380", member.peerScheme(), member.Addr())
}

// ClientURL is the client URL for this member
func (member *Member) ClientURL() string {
	return fmt.Sprintf("%s://%s:2379", member.clientScheme(), member.Addr())
}

func (member *Member) clientScheme() string {
	if member.SecureClient {
		return "https"
	}
	return "http"
}

func (member *Member) ListenClientURL() string {
	return fmt.Sprintf("%s://0.0.0.0:2379", member.clientScheme())
}

func (member *Member) ListenPeerURL() string {
	return fmt.Sprintf("%s://0.0.0.0:2380", member.peerScheme())
}

type MemberSet map[string]*Member

func (memberSet MemberSet) PeerURLPairs() []string {
	urlPairs := make([]string, 0)
	for _, m := range memberSet {
		urlPairs = append(urlPairs, fmt.Sprintf("%s=%s", m.Name, m.PeerURL()))
	}

	return urlPairs
}

func NewMemberSet(ms ...*Member) MemberSet {
	res := MemberSet{}
	for _, m := range ms {
		res[m.Name] = m
	}

	return res
}
