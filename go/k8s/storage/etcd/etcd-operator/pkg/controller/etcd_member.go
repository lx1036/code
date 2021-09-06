package controller

import (
	"fmt"
	"os"
	"time"

	v1 "k8s-lx1036/k8s/storage/etcd/etcd-operator/pkg/apis/etcd.k9s.io/v1"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	EnvOperatorPodName      = "MY_POD_NAME"
	EnvOperatorPodNamespace = "MY_POD_NAMESPACE"
)

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

func (memberSet MemberSet) Add(member *Member) {
	memberSet[member.Name] = member
}

func (memberSet MemberSet) Remove(name string) {
	delete(memberSet, name)
}

func (memberSet MemberSet) Size() int {
	return len(memberSet)
}

// IsEqual INFO: size 必须一样，且 name 必须两者都存在
func (memberSet MemberSet) IsEqual(others MemberSet) bool {
	if memberSet.Size() != others.Size() {
		return false
	}

	for name := range memberSet {
		if _, ok := others[name]; !ok {
			return false
		}
	}

	return true
}

func (memberSet MemberSet) Diff(others MemberSet) MemberSet {
	diff := MemberSet{}
	for name, member := range memberSet {
		if _, ok := others[name]; !ok {
			diff.Add(member)
		}
	}

	return diff
}

func (memberSet MemberSet) ClientURLs() []string {
	clientURLs := make([]string, 0, len(memberSet))
	for _, member := range memberSet {
		clientURLs = append(clientURLs, member.ClientURL())
	}

	return clientURLs
}

// INFO: 这里是直接取值第一个???
func (memberSet MemberSet) PickOne() *Member {
	for _, member := range memberSet {
		return member
	}

	panic("memberSet is empty")
}

func NewMemberSet(ms ...*Member) MemberSet {
	res := MemberSet{}
	for _, m := range ms {
		res[m.Name] = m
	}

	return res
}

func podsToMemberSet(pods []*corev1.Pod, secureClient bool) MemberSet {
	members := MemberSet{}
	for _, pod := range pods {
		members.Add(&Member{
			Name:         pod.Name,
			Namespace:    pod.Namespace,
			SecureClient: secureClient,
		})
	}

	return members
}

func NewMemberAddEvent(memberName string, etcdCluster *v1.EtcdCluster) *corev1.Event {
	event := NewClusterEvent(etcdCluster)
	event.Type = corev1.EventTypeNormal
	event.Reason = "New Member Added"
	event.Message = fmt.Sprintf("New member %s added to cluster", memberName)

	return event
}

func MemberRemoveEvent(memberName string, etcdCluster *v1.EtcdCluster) *corev1.Event {
	event := NewClusterEvent(etcdCluster)
	event.Type = corev1.EventTypeNormal
	event.Reason = "Member Removed"
	event.Message = fmt.Sprintf("Existing member %s removed from the cluster", memberName)

	return event
}

func ReplacingDeadMemberEvent(memberName string, etcdCluster *v1.EtcdCluster) *corev1.Event {
	event := NewClusterEvent(etcdCluster)
	event.Type = corev1.EventTypeNormal
	event.Reason = "Replacing Dead Member"
	event.Message = fmt.Sprintf("The dead member %s is being replaced", memberName)

	return event
}

func NewClusterEvent(etcdCluster *v1.EtcdCluster) *corev1.Event {
	now := time.Now()
	return &corev1.Event{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: etcdCluster.Name + "-",
			Namespace:    etcdCluster.Namespace,
		},
		InvolvedObject: corev1.ObjectReference{
			APIVersion:      v1.SchemeGroupVersion.String(),
			Kind:            v1.EtcdClusterResourceKind,
			Name:            etcdCluster.Name,
			Namespace:       etcdCluster.Namespace,
			UID:             etcdCluster.UID,
			ResourceVersion: etcdCluster.ResourceVersion,
		},
		Source: corev1.EventSource{
			Component: os.Getenv(EnvOperatorPodName),
		},
		// Each cluster event is unique so it should not be collapsed with other events
		FirstTimestamp: metav1.Time{Time: now},
		LastTimestamp:  metav1.Time{Time: now},
		Count:          int32(1),
	}
}
