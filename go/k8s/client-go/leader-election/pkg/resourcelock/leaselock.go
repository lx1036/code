package resourcelock

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	coordinationv1 "k8s.io/api/coordination/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	coordinationv1client "k8s.io/client-go/kubernetes/typed/coordination/v1"
)

type LeaseLock struct {

	// LeaseMeta should contain a Name and a Namespace of a
	// LeaseMeta object that the LeaderElector will attempt to lead.
	LeaseMeta  metav1.ObjectMeta
	Client     coordinationv1client.LeasesGetter
	LockConfig ResourceLockConfig
	lease      *coordinationv1.Lease
}

func LeaseSpecToLeaderElectionRecord(spec *coordinationv1.LeaseSpec) *LeaderElectionRecord {
	var r LeaderElectionRecord
	if spec.HolderIdentity != nil {
		r.HolderIdentity = *spec.HolderIdentity
	}
	if spec.LeaseDurationSeconds != nil {
		r.LeaseDurationSeconds = int(*spec.LeaseDurationSeconds)
	}
	if spec.LeaseTransitions != nil {
		r.LeaderTransitions = int(*spec.LeaseTransitions)
	}
	if spec.AcquireTime != nil {
		r.AcquireTime = metav1.Time{Time: spec.AcquireTime.Time}
	}
	if spec.RenewTime != nil {
		r.RenewTime = metav1.Time{Time: spec.RenewTime.Time}
	}

	return &r
}

func (leaseLock *LeaseLock) Get(ctx context.Context) (*LeaderElectionRecord, []byte, error) {
	var err error
	leaseLock.lease, err = leaseLock.Client.Leases(leaseLock.LeaseMeta.Namespace).Get(ctx, leaseLock.LeaseMeta.Name, metav1.GetOptions{})
	if err != nil {
		return nil, nil, err
	}

	record := LeaseSpecToLeaderElectionRecord(&leaseLock.lease.Spec)
	recordByte, err := json.Marshal(*record)
	if err != nil {
		return nil, nil, err
	}

	return record, recordByte, nil
}

func (leaseLock *LeaseLock) Create(ctx context.Context, ler LeaderElectionRecord) error {
	var err error
	leaseLock.lease, err = leaseLock.Client.Leases(leaseLock.LeaseMeta.Namespace).Create(ctx, &coordinationv1.Lease{
		ObjectMeta: metav1.ObjectMeta{
			Name:      leaseLock.LeaseMeta.Name,
			Namespace: leaseLock.LeaseMeta.Namespace,
		},
		Spec: LeaderElectionRecordToLeaseSpec(&ler),
	}, metav1.CreateOptions{})
	return err
}

func (leaseLock *LeaseLock) Update(ctx context.Context, ler LeaderElectionRecord) error {
	if leaseLock.lease == nil {
		return errors.New("lease not initialized, call get or create first")
	}

	leaseLock.lease.Spec = LeaderElectionRecordToLeaseSpec(&ler)

	lease, err := leaseLock.Client.Leases(leaseLock.LeaseMeta.Namespace).Update(ctx, leaseLock.lease, metav1.UpdateOptions{})
	if err != nil {
		return err
	}

	leaseLock.lease = lease
	return nil
}

func (leaseLock *LeaseLock) Describe() string {
	return fmt.Sprintf("%v/%v", leaseLock.LeaseMeta.Namespace, leaseLock.LeaseMeta.Name)
}

// Identity returns the Identity of the lock
func (leaseLock *LeaseLock) Identity() string {
	return leaseLock.LockConfig.Identity
}

func LeaderElectionRecordToLeaseSpec(leaderElectionRecord *LeaderElectionRecord) coordinationv1.LeaseSpec {
	leaseDurationSeconds := int32(leaderElectionRecord.LeaseDurationSeconds)
	leaseTransitions := int32(leaderElectionRecord.LeaderTransitions)
	return coordinationv1.LeaseSpec{
		HolderIdentity:       &leaderElectionRecord.HolderIdentity,
		LeaseDurationSeconds: &leaseDurationSeconds,
		AcquireTime:          &metav1.MicroTime{Time: leaderElectionRecord.AcquireTime.Time},
		RenewTime:            &metav1.MicroTime{Time: leaderElectionRecord.RenewTime.Time},
		LeaseTransitions:     &leaseTransitions,
	}
}
