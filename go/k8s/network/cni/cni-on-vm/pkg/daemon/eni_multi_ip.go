package daemon

import (
	"context"
	"fmt"
	"k8s-lx1036/k8s/network/cni/cni-on-vm/pkg/ipam"
	"k8s-lx1036/k8s/network/cni/cni-on-vm/pkg/utils/types"
	"k8s.io/klog/v2"
	"net"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// trunk 模式：@see https://github.com/AliyunContainerService/terway/blob/main/docs/terway-trunk.md
// 1. 固定IP
// 2. Pod 配置独立 vSwitch、安全组
// 3. 可为一组 Pod或 namespace 进行配置

const (
	maxEniOperating = 3
	maxENI          = 10 // 一个 vm 最多有 10 个弹性网卡，且每个弹性网卡最多 10 个 IPs
)

type networkContext struct {
	context.Context
	resources  []types.ResourceItem
	pod        *types.PodInfo
	k8sService K8sService
}

type eniIPResourceManager struct {
	trunkENI *types.ENI
	pool     *SimpleObjectPool
}

func newENIIPResourceManager(poolConfig *ResourceConfig, ecs ipam.API, k8s K8sService, allocatedResources map[string]resourceManagerInitItem) (*eniIPResourceManager, error) {
	factory := NewENIIPFactory(poolConfig, ecs)

	p, err := NewSimpleObjectPool(PoolConfig{
		Name:     "eniIP",
		Type:     "eniIP",
		MaxIdle:  poolConfig.MaxPoolSize,
		MinIdle:  poolConfig.MinPoolSize,
		Factory:  factory,
		Capacity: capacity,
	}, ecs, nil)
	if err != nil {
		return nil, err
	}

	var trunkENI *types.ENI

	mgr := &eniIPResourceManager{
		pool:     p,
		trunkENI: trunkENI,
	}

	return mgr, nil
}

func (m *eniIPResourceManager) Allocate(ctx *networkContext, id string) (types.NetworkResource, error) {
	return m.pool.Acquire(ctx, id, podInfoKey(ctx.pod.Namespace, ctx.pod.Name))
}

func (m *eniIPResourceManager) Release(ctx *networkContext, resItem types.ResourceItem) error {
	if ctx != nil && ctx.pod != nil {
		return m.pool.ReleaseWithReservation(resItem.ID, ctx.pod.IPStickTime)
	}
	return m.pool.Release(resItem.ID)
}

type eniIPFactory struct {
	sync.RWMutex

	name        string
	eniMaxIP    int
	enableTrunk bool

	enis       []*ENI // 该机器上所有虚拟网卡
	eniFactory *eniFactory

	ipResultChan   chan *ENIIP
	eniOperateChan chan struct{}
	maxENIChan     chan struct{}

	// metrics
	metricENICount prometheus.Gauge
}

func NewENIIPFactory(poolConfig *ResourceConfig, ecs ipam.API) *eniIPFactory {
	factory := &eniIPFactory{
		name:        "eniIP",
		eniFactory:  newENIFactory(poolConfig, ecs),
		enis:        make([]*ENI, 0),
		enableTrunk: poolConfig.EnableENITrunking,

		ipResultChan:   make(chan *ENIIP, maxIPBacklog),
		eniOperateChan: make(chan struct{}, maxEniOperating),
		maxENIChan:     make(chan struct{}, maxENI),
	}

	return factory
}

func (f *eniIPFactory) Reconcile() {
	// check ENI SecurityGroup
}

// ListResource load all eni info from metadata
func (f *eniIPFactory) ListResource() (map[string]types.NetworkResource, error) {
	return nil, nil
}

// Create call IP API to allocate next count ip for current EIP. 为所有 ENI 都分配 count IP
func (f *eniIPFactory) Create(count int) ([]types.NetworkResource, error) {
	var (
		waiting  int
		ipResult []types.NetworkResource
	)

	for ; waiting < count; waiting++ {
		enis := f.getENIs()
		for _, eni := range enis {
			if eni.getIPCountLocked() < f.eniMaxIP {
				select {
				case eni.ipBacklog <- struct{}{}:
				default:
					continue
				}

				eni.pending++
			}
		}
	}

	// 如果还有 remaining IP 需要分配，则需要新建一个网卡
	remaining := count - waiting
	if remaining > f.eniMaxIP {
		remaining = f.eniMaxIP
	}
	if remaining > maxIPBacklog {
		remaining = maxIPBacklog
	}
	if remaining > 0 {
		_, err := f.createENIAsync(remaining)
		if err == nil {
			waiting += remaining
		} else {
			klog.Errorf(fmt.Sprintf("create ENI async err:%v", err))
		}
	}

	// no ip has been allocated
	if waiting == 0 {
		return ipResult, fmt.Errorf("no ip has been allocated")
	}

	for ; waiting > 0; waiting-- {
		result := <-f.ipResultChan
		if result.ENIIP == nil || result.err != nil {
			continue
		}

		for _, eni := range f.enis {
			if eni.ENI != nil && eni.ENI.MAC == result.ENI.MAC {
				eni.pending--
				eni.ips = append(eni.ips, result)
				ipResult = append(ipResult, result.ENIIP)
			}
		}
	}

	if len(ipResult) == 0 {
		return nil, fmt.Errorf("error allocate ip address")
	}

	return ipResult, nil
}

func (f *eniIPFactory) Dispose() error {

}

func (f *eniIPFactory) getENIs() []*ENI {

}

func (f *eniIPFactory) createENIAsync(initIPs int) (*ENI, error) {
	eni := &ENI{
		ENI:                    nil,
		ips:                    make([]*ENIIP, 0),
		pending:                0,
		ipBacklog:              nil,
		ecs:                    f.eniFactory.ecs,
		done:                   nil,
		ipAllocInhibitExpireAt: time.Time{},
	}

	select {
	case f.maxENIChan <- struct{}{}:
		select {
		case f.eniOperateChan <- struct{}{}:
		default:
			<-f.maxENIChan
			return nil, fmt.Errorf("trigger ENI throttle, max operating concurrent: %v", maxEniOperating)
		}
		go f.createENI(eni, eni.pending)
	default:
		return nil, fmt.Errorf("max ENI exceeded")
	}

	f.Lock()
	f.enis = append(f.enis, eni)
	f.Unlock()

	// metric
	f.metricENICount.Inc()
	return eni, nil
}

func (f *eniIPFactory) createENI(eni *ENI, ipCount int) {
	// create ENI
	rawEni, err := f.eniFactory.CreateWithIPCount(ipCount, false)

	// eni operate finished
	<-f.eniOperateChan

	var ipv4s []net.IP
	if err != nil || len(rawEni) != 1 {
		// create eni failed, put quota back
		<-f.maxENIChan
	} else {
		eni.ENI = rawEni[0].(*types.ENI)
		ipv4s, _, err = f.eniFactory.ecs.GetENIIPs(context.Background(), eni.MAC)
	}

	for _, ipv4 := range ipv4s {
		eniIP := &types.ENIIP{
			ENI: eni.ENI,
			IPSet: types.IPSet{
				IPv4: ipv4,
			},
		}
		f.ipResultChan <- &ENIIP{
			ENIIP: eniIP,
			err:   nil,
		}
	}

	go eni.allocateWorker(f.ipResultChan)
}
