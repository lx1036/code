package daemon

import (
	"context"
	"errors"
	"fmt"
	"k8s-lx1036/k8s/network/cni/vpc-cni/pkg/utils/metric"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
	"sync"
	"time"

	"k8s-lx1036/k8s/network/cni/vpc-cni/pkg/ipam"
	"k8s-lx1036/k8s/network/cni/vpc-cni/pkg/utils/types"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	CheckIdleInterval = 2 * time.Minute

	defaultPoolBackoff = 30 * time.Second

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
	Name     string
	Type     string
	Factory  *eniIPFactory
	MinIdle  int
	MaxIdle  int
	Capacity int
}

type SimpleObjectPool struct {
	lock sync.Mutex

	name         string
	eniIPFactory *eniIPFactory

	idle        *PriorityQueue
	inuse       map[string]poolItem
	invalid     map[string]poolItem
	maxIdle     int
	minIdle     int
	capacity    int
	backoffTime time.Duration

	metricIdle     prometheus.Gauge
	metricTotal    prometheus.Gauge
	metricDisposed prometheus.Counter

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
		name:         cfg.Name,
		eniIPFactory: cfg.Factory,

		inuse:   make(map[string]poolItem),
		idle:    NewPriorityQueue(),
		invalid: make(map[string]poolItem),

		maxIdle:     cfg.MaxIdle,
		minIdle:     cfg.MinIdle,
		capacity:    cfg.Capacity,
		notifyCh:    make(chan interface{}, 1),
		tokenCh:     make(chan struct{}, cfg.Capacity),
		backoffTime: defaultPoolBackoff,

		metricIdle:     metric.ResourcePoolIdle.WithLabelValues(cfg.Name, cfg.Type, fmt.Sprint(cfg.Capacity), fmt.Sprint(cfg.MaxIdle), fmt.Sprint(cfg.MinIdle)),
		metricTotal:    metric.ResourcePoolTotal.WithLabelValues(cfg.Name, cfg.Type, fmt.Sprint(cfg.Capacity), fmt.Sprint(cfg.MaxIdle), fmt.Sprint(cfg.MinIdle)),
		metricDisposed: metric.ResourcePoolDisposed.WithLabelValues(cfg.Name, cfg.Type, fmt.Sprint(cfg.Capacity), fmt.Sprint(cfg.MaxIdle), fmt.Sprint(cfg.MinIdle)),
	}

	// not use main ENI for ENI multiple ip allocate
	ctx := context.Background()
	enis, err := ecs.GetAttachedENIs(ctx, false)
	if err != nil {
		return nil, fmt.Errorf("get attach ENI on pool init err:%v", err)
	}
	if pool.eniIPFactory.enableTrunk {
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
		pool.eniIPFactory.enis = append(pool.eniIPFactory.enis, poolENI)
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
		case pool.eniIPFactory.maxENIChan <- struct{}{}:
		default:
			klog.Warningf("exist enis already over eni limits, maxENI config will not be available")
		}

		go poolENI.allocateWorker(pool.eniIPFactory.ipResultChan)
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

	tokenCount := p.capacity - p.sizeLocked() // capacity 减去已经使用的，preload 先补位
	for i := 0; i < tokenCount; i++ {
		p.tokenCh <- struct{}{} // @see checkInsufficient()
	}

	return nil
}

func (p *SimpleObjectPool) sizeLocked() int {
	return p.idle.Size() + len(p.inuse) + len(p.invalid)
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
		case <-tick: // 每次 2min ~ 4min 一次循环
			p.checkIdle()         // 多了则回收
			p.checkInsufficient() // 少了则创建
		case <-reconcileTick: // 每次 1h ~ 2h 一次循环
			p.eniIPFactory.Reconcile()
		case <-p.notifyCh: // 立即检查
			p.checkIdle()
			p.checkInsufficient()
		}
	}
}

func (p *SimpleObjectPool) notify() {
	select {
	case p.notifyCh <- true:
	default:
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

	p.inuse[key] = poolItem{res: eniIP, idempotentKey: key}

	p.metricIdle.Inc()
}

// INFO: 回收多余的 idle ENIIP, idle 数量在 [minIdle, maxIdle]，如[0,10]，多余的部分则根据过期时间回收 ENIIP 资源(用后端 api，释放该弹性网卡 ENI 的 IP)
func (p *SimpleObjectPool) checkIdle() {
	for {
		item := p.peekIdleExpired()
		if item == nil {
			break
		}

		err := p.eniIPFactory.Dispose(item.res)
		if err == nil {
			klog.Infof(fmt.Sprintf("dispose ENIIP xxx"))

			p.metricIdle.Dec()
			p.metricTotal.Dec()
			p.metricDisposed.Inc()

			p.tokenCh <- struct{}{}
			p.backoffTime = defaultPoolBackoff
		} else {
			klog.Errorf(fmt.Sprintf("dispose ENIIP xxx err:%v", err))
			p.AddIdle(item.res)
			p.backoffTime = p.backoffTime * 2
			time.Sleep(p.backoffTime)
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

// INFO: 创建不足的 idle ENIIP, idle 数量在 [minIdle, maxIdle]，如[5,10]，不足的部分则调用后端 api 来分配 ENIIP
func (p *SimpleObjectPool) checkInsufficient() {
	addition := p.minIdle - p.idle.Size()
	if addition > (p.capacity - p.sizeLocked()) {
		addition = p.capacity - p.sizeLocked()
	}
	if addition <= 0 {
		return
	}

	var tokenAcquired int
	for i := 0; i < addition; i++ {
		// pending resources
		select {
		case <-p.tokenCh:
			tokenAcquired++
		default:
			continue
		}
	}
	if tokenAcquired <= 0 {
		return
	}

	res, err := p.eniIPFactory.Create(tokenAcquired) // 可能后端创建不了足够的 tokenAcquired 个 ENIIP
	if err != nil {
		klog.Errorf(fmt.Sprintf("error create from factory: %v", err))
		p.backoffTime = p.backoffTime * 2
		time.Sleep(p.backoffTime)
		return
	}
	if tokenAcquired == len(res) {
		p.backoffTime = defaultPoolBackoff
	}

	for _, eniIP := range res {
		p.AddIdle(eniIP)
		tokenAcquired--
	}
	for i := 0; i < tokenAcquired; i++ {
		// release token
		p.tokenCh <- struct{}{}
	}

	if tokenAcquired != 0 {
		p.notify()
	}
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
		res, err := p.eniIPFactory.Create(1)
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

func (p *SimpleObjectPool) Release(resID string) error {
	return p.ReleaseWithReservation(resID, 0)
}

func (p *SimpleObjectPool) ReleaseWithReservation(resID string, reservation time.Duration) error {
	p.lock.Lock()
	defer p.lock.Unlock()

	res, ok := p.inuse[resID]
	if !ok {
		klog.Errorf(fmt.Sprintf("release %s err %v", resID, ErrInvalidState))
		return ErrInvalidState
	}

	delete(p.inuse, resID)

	reserveTo := time.Now()
	if reservation > 0 {
		reserveTo = reserveTo.Add(reservation)
	}

	p.idle.Push(&poolItem{res: res.res, reservation: reserveTo})
	p.metricIdle.Inc()
	p.notify() // 立即检查 idle
	return nil
}
