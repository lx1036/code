package daemon

import (
	"context"
	"errors"
	"fmt"
	"k8s-lx1036/k8s/network/cni/cni-on-vm/pkg/metric"
	"k8s.io/apimachinery/pkg/util/wait"
	"sync"
	"time"

	"k8s-lx1036/k8s/network/cni/cni-on-vm/pkg/ipam"
	"k8s-lx1036/k8s/network/cni/cni-on-vm/pkg/types"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	CheckIdleInterval = 2 * time.Minute
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

type eniIPFactory struct {
	name string
}

// ListResource load all eni info from metadata
func (f *eniIPFactory) ListResource() (map[string]types.NetworkResource, error) {

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
	lock sync.Mutex

	idle     *priorityQueue
	inuse    map[string]poolItem
	invalid  map[string]poolItem
	maxIdle  int
	minIdle  int
	capacity int

	metricIdle  prometheus.Gauge
	metricTotal prometheus.Gauge
}

func NewSimpleObjectPool(cfg Config, ecs ipam.API, allocatedResources map[string]types.ResourceManagerInitItem) (*SimpleObjectPool, error) {
	if cfg.MinIdle > cfg.MaxIdle {
		return nil, ErrInvalidArguments
	}

	if cfg.MaxIdle > cfg.Capacity {
		return nil, ErrInvalidArguments
	}

	pool := &SimpleObjectPool{
		name:    cfg.Name,
		factory: cfg.Factory,

		inuse:   make(map[string]poolItem),
		idle:    NewPriorityQueue(),
		invalid: make(map[string]poolItem),

		maxIdle:     cfg.MaxIdle,
		minIdle:     cfg.MinIdle,
		capacity:    cfg.Capacity,
		notifyCh:    make(chan interface{}, 1),
		tokenCh:     make(chan struct{}, cfg.Capacity),
		backoffTime: defaultPoolBackoff,

		metricIdle:  metric.ResourcePoolIdle.WithLabelValues(cfg.Name, cfg.Name, fmt.Sprint(cfg.Capacity), fmt.Sprint(cfg.MaxIdle), fmt.Sprint(cfg.MinIdle)),
		metricTotal: metric.ResourcePoolTotal.WithLabelValues(cfg.Name, cfg.Name, fmt.Sprint(cfg.Capacity), fmt.Sprint(cfg.MaxIdle), fmt.Sprint(cfg.MinIdle)),
	}

	// not use main ENI for ENI multiple ip allocate
	ctx := context.Background()
	enis, err := ecs.GetAttachedENIs(ctx, false)

	for _, eni := range enis {
		ipv4s, _, err := ecs.GetENIIPs(ctx, eni.ID) // 使用 ENI ID 查询 ENI IP info

		poolENI := &ENI{
			ENI:       eni,
			ips:       []*ENIIP{},
			ecs:       ecs,
			ipBacklog: make(chan struct{}, maxIPBacklog),
			done:      make(chan struct{}, 1),
		}

		for _, ip := range ipv4s {
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
				pool.AddInuse(eniIP, types.PodInfoKey(res.PodInfo.Namespace, res.PodInfo.Name))
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

func (p *SimpleObjectPool) startCheckIdleTicker() {
	// 相比于 time.Tick()，有了 jitter
	tick := make(chan struct{})
	go wait.JitterUntil(func() {
		tick <- struct{}{}
	}, CheckIdleInterval, 0.2, true, wait.NeverStop) // 每次 2min ~ 4min 一次循环
	reconcileTick := make(chan struct{})
	go wait.JitterUntil(func() {
		reconcileTick <- struct{}{}
	}, time.Hour, 0.2, true, wait.NeverStop) // 每次 1h ~ 2h 一次循环

	for {
		select {
		case <-tick:
			p.checkResync() // make sure pool is synced
			p.checkIdle()
			p.checkInsufficient()
		case <-reconcileTick:
			p.factory.Reconcile()
		case <-p.notifyCh:
			p.checkIdle()
			p.checkInsufficient()
		}
	}
}

func (p *SimpleObjectPool) AddIdle(eniIP *types.ENIIP) {
	p.lock.Lock()
	defer p.lock.Unlock()

	p.idle.Push(&poolItem{res: eniIP, reservation: time.Now()})

	p.metricTotal.Inc()
	p.metricIdle.Inc()
}

func (p *SimpleObjectPool) AddInuse(eniIP *types.ENIIP, key string) {
	p.lock.Lock()
	defer p.lock.Unlock()

	p.inuse[key] = poolItem{res: eniIP, key: key}

	p.metricIdle.Inc()
}

// 检查过期的 ENIIP 资源，然后调用后端 api，释放该弹性网卡 ENI 的 IP
func (p *SimpleObjectPool) checkIdle() {
	for {
		item := p.peekIdleExpired()
		if item == nil {
			break
		}

	}
}

func (p *SimpleObjectPool) peekIdleExpired() *poolItem {
	p.lock.Lock()
	defer p.lock.Unlock()

	if !p.tooManyIdleLocked() {
		return nil
	}

	item := p.idle.Peek()
	if item == nil {
		return nil
	}

	if item.reservation.After(time.Now()) {
		return nil
	}
	return p.idle.Pop()
}

func (p *SimpleObjectPool) tooManyIdleLocked() bool {
	return p.idle.Size() > p.maxIdle || (p.idle.Size() > 0 && p.sizeLocked() > p.capacity)
}

func (p *SimpleObjectPool) sizeLocked() int {
	return p.idle.Size() + len(p.inuse) + len(p.invalid)
}

// @see https://github.com/kubernetes/kubernetes/blob/v1.23.5/staging/src/k8s.io/client-go/util/workqueue/delaying_queue.go
// @see https://github.com/AliyunContainerService/terway/blob/main/pkg/pool/queue.go

type eniIPItem struct {
	res         *types.ENIIP
	expiration time.Time
	key         string
}

// 最小/大堆实现 priority queue
type eniIPPriorityQueue []*eniIPItem

func (q *eniIPPriorityQueue) Less(i, j int) bool {
	return q.items[i].reservation.Before(q.items[j].reservation)
}

func (q *eniIPPriorityQueue) Swap(i, j int) {
	q.items[i], q.items[j] = q.items[j], q.items[i]
}

func (q *eniIPPriorityQueue) Push(item *poolItem) {
	q.items = append(q.items, item)

	// bubble up
	index := len(q.items) - 1
	for index > 0 {
		parent := (index - 1) / 2
		if !q.items[index].less(q.items[parent]) {
			break
		}
		q.Swap(index, parent)
		index = parent
	}

}

func (q *eniIPPriorityQueue) Peek() *poolItem {
	return q.items[0]
}

func (q *eniIPPriorityQueue) Pop() *poolItem {
	if q.size == 0 {
		return nil
	}

	item := q.items[0]
	q.items[0] = q.items[q.size-1]
	q.size--
	q.bubbleDowm(0)
	return item
}

func (q *eniIPPriorityQueue) Size() int {
	return len(*q)
}
