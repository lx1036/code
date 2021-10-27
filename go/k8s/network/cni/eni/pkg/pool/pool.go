package pool

import (
	"context"
	"time"

	"k8s-lx1036/k8s/network/cni/eni/pkg/types"
)

// ObjectPool object pool interface
type ObjectPool interface {
	Acquire(ctx context.Context, resID, idempotentKey string) (types.NetworkResource, error)
	ReleaseWithReservation(resID string, reservation time.Duration) error
	Release(resID string) error
	AcquireAny(ctx context.Context, idempotentKey string) (types.NetworkResource, error)
	Stat(resID string) (types.NetworkResource, error)
	GetName() string
	//tracing.ResourceMappingHandler
}

// ObjectFactory interface of network resource object factory
type ObjectFactory interface {
	// Create res with count
	Create(count int) ([]types.NetworkResource, error)
	Dispose(types.NetworkResource) error
	ListResource() (map[string]types.NetworkResource, error)
	Check(types.NetworkResource) error
	// Reconcile run periodicity
	Reconcile()
}
