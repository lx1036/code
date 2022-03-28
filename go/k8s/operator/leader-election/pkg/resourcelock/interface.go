package resourcelock

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	LeaderElectionRecordAnnotationKey = "control-plane.alpha.lx1036/leader"
	EndpointsResourceLock             = "endpoints"
	ConfigMapsResourceLock            = "configmaps"
	LeasesResourceLock                = "leases"
	EndpointsLeasesResourceLock       = "endpointsleases"
	ConfigMapsLeasesResourceLock      = "configmapsleases"
)

type Interface interface {
	Describe() string

	Get(ctx context.Context) (*LeaderElectionRecord, []byte, error)

	Create(ctx context.Context, ler LeaderElectionRecord) error

	Identity() string

	Update(ctx context.Context, ler LeaderElectionRecord) error
}

// LeaderElectionRecord is the record that is stored in the leader election annotation.
// This information should be used for observational purposes only.
// @see staging/src/k8s.io/api/coordination/v1/types.go LeaseSpec
type LeaderElectionRecord struct {
	HolderIdentity       string `json:"holderIdentity"`
	LeaseDurationSeconds int    `json:"leaseDurationSeconds"`

	// leader首次acquire时间
	AcquireTime metav1.Time `json:"acquireTime"`
	// leader重新renew时间
	RenewTime metav1.Time `json:"renewTime"`

	// leader切换次数
	LeaderTransitions int `json:"leaderTransitions"`
}

type ResourceLockConfig struct {
	// Identity is the unique string identifying a lease holder across
	// all participants in an election.
	Identity string
	// EventRecorder is optional.
	//EventRecorder EventRecorder
}
