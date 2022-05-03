package daemon

import (
	"context"
	"errors"
	"fmt"
	"k8s-lx1036/k8s/network/cni/cni-on-vm/pkg/metric"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
	"sync"
	"time"

	"k8s-lx1036/k8s/network/cni/cni-on-vm/pkg/ipam"
	"k8s-lx1036/k8s/network/cni/cni-on-vm/pkg/types"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	CheckIdleInterval = 2 * time.Minute

	defaultPoolBackoff = 1 * time.Minute

	// 每一个弹性网卡最多 10 个 IP
	maxIPBacklog = 10
)

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

type PoolConfig struct {
	Name        string
	Type        string
	Factory     *eniIPFactory
	Initializer Initializer
	MinIdle     int
	MaxIdle     int
	Capacity    int
}

type SimpleObjectPool struct {
	lock sync.Mutex

	name    string
	factory *eniIPFactory

	idle        *PriorityQueue
	inuse       map[string]poolItem
	invalid     map[string]poolItem
	maxIdle     int
	minIdle     int
	capacity    int
	backoffTime time.Duration

	metricIdle  prometheus.Gauge
	metricTotal prometheus.Gauge

	notifyCh chan interface{}
	// concurrency to create resource. tokenCh = capacity - (idle + inuse + dispose)
	tokenCh chan struct{}
}

func NewSimpleObjectPool(cfg PoolConfig, ecs ipam.API, allocatedResources map[string]types.ResourceManagerInitItem) (*SimpleObjectPool, error) {
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
	if err != nil {
		return nil, fmt.Errorf("get attach ENI on pool init err:%v", err)
	}
	if pool.factory.enableTrunk {
		// TODO: trunk 模式 @see https://github.com/AliyunContainerService/terway/blob/main/docs/terway-trunk.md
	}
	// 每一个 eni 网卡开启一个 goroutine，然后为该网卡分配一定数量的 IP。
	// 类似于 hashicorp/raft leader 为每一个 follower 开启一个 goroutine 来 replicate/heartbeat log。
	for _, eni := range enis {
		ipv4s, _, err := ecs.GetENIIPs(ctx, eni.ID) // 使用 ENI ID 查询 ENI IP info
		if err != nil {
			klog.Errorf(fmt.Sprintf("get ips for eni %s err:%v", eni.ID, err))
			continue
		}
		poolENI := &ENI{
			ENI:       eni,
			ips:       make([]*ENIIP, 0),
			ecs:       ecs,
			ipBacklog: make(chan struct{}, maxIPBacklog),
			done:      make(chan struct{}, 1),
		}
		pool.factory.enis = append(pool.factory.enis, poolENI)
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

		select {
		case pool.factory.maxENIChan <- struct{}{}:
		default:
			klog.Warningf("exist enis already over eni limits, maxENI config will not be available")
		}

		go poolENI.allocateWorker(pool.factory.ipResultChan)
	}

	if err := pool.preload(); err != nil {
		return nil, err
	}

	go pool.startCheckIdleTicker()

	return pool, nil
}

func (p *SimpleObjectPool) preload() error {
	p.lock.Lock()
	defer p.lock.Unlock()

	tokenCount := p.capacity - p.sizeLocked()
	for i := 0; i < tokenCount; i++ {
		p.tokenCh <- struct{}{}
	}

	return nil
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

func (p *SimpleObjectPool) Acquire(ctx context.Context, resID, idempotentKey string) (types.NetworkResource, error) {
	p.lock.Lock()
	if resItem, ok := p.inuse[resID]; ok && resItem.idempotentKey == idempotentKey {
		p.lock.Unlock()
		return resItem.res, nil
	}

	if p.idle.Size() > 0 {
		var item *poolItem
		if len(resID) > 0 {
			item = p.idle.Rob(resID)
			if item == nil {
				item = p.idle.Pop()
			}
		}
		res := item.res
		p.inuse[res.GetResourceID()] = poolItem{
			res:           res,
			reservation:   time.Time{},
			idempotentKey: idempotentKey,
		}
		p.lock.Unlock()
		p.metricIdle.Dec()
		p.notify()
		klog.Infof(fmt.Sprintf("acquire ip resource xxx"))
		return res, nil
	}

	size := p.sizeLocked()
	if size >= p.capacity {
		p.lock.Unlock()
		klog.Infof(fmt.Sprintf("acquire (expect %s), size %d, capacity %d: return err %v", resID, size,
			p.capacity, ErrNoAvailableResource))

		return nil, ErrNoAvailableResource
	}

	p.lock.Unlock()

	select {
	case <-p.tokenCh: // call IP API for create ip for current ENI
		res, err := p.factory.Create(1)
		if err != nil || len(res) == 0 {
			p.tokenCh <- struct{}{}
			return nil, fmt.Errorf("error create from factory: %v", err)
		}

		klog.Infof(fmt.Sprintf("call IP API for create ip xxx"))
		p.AddInuse(res[0], idempotentKey)
		return res[0], nil
	case <-ctx.Done():
		return nil, ErrContextDone
	}
}

func (p *SimpleObjectPool) notify() {
	select {
	case p.notifyCh <- true:
	default:
	}
}
