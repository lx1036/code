package pool

import (
	"context"
	"errors"
	"k8s-lx1036/k8s/network/cni/cni-on-vm/pkg/ipam"
	"sync"
	"time"

	"k8s-lx1036/k8s/network/cni/cni-on-vm/pkg/types"
)

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

var (
	ErrNoAvailableResource = errors.New("no available resource")
	ErrInvalidState        = errors.New("invalid state")
	ErrNotFound            = errors.New("not found")
	ErrContextDone         = errors.New("context done")
	ErrInvalidArguments    = errors.New("invalid arguments")
)

type ENIIP struct {
	*types.ENIIP
	err error
}

type ENI struct {
	lock sync.Mutex

	*types.ENI
	ips       []*ENIIP
	pending   int
	ipBacklog chan struct{}
	ecs       ipam.API
	done      chan struct{}
	// Unix timestamp to mark when this ENI can allocate Pod IP.
	ipAllocInhibitExpireAt time.Time
}

type Config struct {
	Name        string
	Type        string
	Factory     ObjectFactory
	Initializer Initializer
	MinIdle     int
	MaxIdle     int
	Capacity    int
}

type SimpleObjectPool struct {
}

func NewSimpleObjectPool(cfg Config, ecs ipam.API) (*SimpleObjectPool, error) {
	if cfg.MinIdle > cfg.MaxIdle {
		return nil, ErrInvalidArguments
	}

	if cfg.MaxIdle > cfg.Capacity {
		return nil, ErrInvalidArguments
	}

	pool := &SimpleObjectPool{
		name:        cfg.Name,
		factory:     cfg.Factory,
		inuse:       make(map[string]poolItem),
		idle:        newPriorityQueue(),
		invalid:     make(map[string]poolItem),
		maxIdle:     cfg.MaxIdle,
		minIdle:     cfg.MinIdle,
		capacity:    cfg.Capacity,
		notifyCh:    make(chan interface{}, 1),
		tokenCh:     make(chan struct{}, cfg.Capacity),
		backoffTime: defaultPoolBackoff,
	}

	// not use main ENI for ENI multiple ip allocate
	ctx := context.Background()
	enis, err := ecs.GetAttachedENIs(ctx, false)

	for _, eni := range enis {
		ipv4s, _, err := ecs.GetENIIPs(ctx, eni.ID) // 使用 ENI ID 查询 ENI 信息

		poolENI := &ENI{
			ENI:       eni,
			ips:       []*ENIIP{},
			ecs:       ecs,
			ipBacklog: make(chan struct{}, maxIPBacklog),
			done:      make(chan struct{}, 1),
		}

		for _, ipv4 := range ipv4s {
			eniIP := &types.ENIIP{
				ENI:   eni,
				IPSet: types.IPSet{IPv4: ip},
			}

			poolENI.ips = append(poolENI.ips, &ENIIP{
				ENIIP: eniIP,
			})

			res, ok := allocatedResources[eniIP.GetResourceID()]
			if !ok {
				pool.AddIdle(eniIP)
			} else {
				pool.AddInuse(eniIP, podInfoKey(res.podInfo.Namespace, res.podInfo.Name))
			}
		}

		go poolENI.allocateWorker(factory.ipResultChan)
	}

	if err := pool.preload(); err != nil {
		return nil, err
	}

	go pool.startCheckIdleTicker()

	return pool, nil
}
