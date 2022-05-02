package daemon

import (
	"context"
	"fmt"
	"sync"
	"time"

	"k8s-lx1036/k8s/network/cni/cni-on-vm/pkg/ipam"
	"k8s-lx1036/k8s/network/cni/cni-on-vm/pkg/pool"
	"k8s-lx1036/k8s/network/cni/cni-on-vm/pkg/types"
)

type eniFactory struct {
	sync.RWMutex

	name                   string
	switches               []string
	eniTags                map[string]string
	securityGroup          string
	instanceID             string
	ecs                    ipam.API
	vswitchIPCntMap        map[string]int
	tsExpireAt             time.Time
	vswitchSelectionPolicy string
}

func newENIFactory(poolConfig *ResourceConfig, ecs ipam.API) *eniFactory {

	return &eniFactory{
		name:                   factoryNameENI,
		switches:               poolConfig.VSwitch,
		eniTags:                poolConfig.ENITags,
		securityGroup:          poolConfig.SecurityGroup,
		instanceID:             poolConfig.InstanceID,
		ecs:                    ecs,
		vswitchIPCntMap:        make(map[string]int),
		vswitchSelectionPolicy: poolConfig.VSwitchSelectionPolicy,
	}
}

func (f *eniFactory) Create(int) ([]types.NetworkResource, error) {
	return f.CreateWithIPCount(1, false)
}

// CreateWithIPCount 在当前 vm 上，创建 ENI 弹性网卡资源
func (f *eniFactory) CreateWithIPCount(count int, trunk bool) ([]types.NetworkResource, error) {

	eni, err := f.ecs.AllocateENI(context.Background(), vSwitches[0], f.securityGroup, f.instanceID, trunk, count, tags)
	if err != nil {
		return nil, err
	}
	return []types.NetworkResource{eni}, nil
}

// ENI 每一个弹性网卡
type ENI struct {
	lock sync.Mutex

	*types.ENI
	ips     []*ENIIP
	pending int

	// @see eniIPFactory.submit(), make(chan struct{}, 10)
	ipBacklog chan struct{}

	ecs  ipam.API
	done chan struct{}
	// Unix timestamp to mark when this ENI can allocate Pod IP.
	ipAllocInhibitExpireAt time.Time
}

// 调用后端 API 为 ENI allocate 多个 IPs
func (eni *ENI) allocateWorker(resultChan chan<- *ENIIP) {
	for {
		toAllocate := 0
		select {
		case <-eni.done:
			return
		case <-eni.ipBacklog:
			toAllocate = 1
		}

		// wait 300ms for aggregation the cni request
		time.Sleep(300 * time.Millisecond)

	popAll:
		for {
			select {
			case <-eni.ipBacklog:
				toAllocate++
			default:
				break popAll
			}
		}

		ips, err := eni.ecs.AssignNIPsForENI(context.TODO(), eni.ENI.ID, eni.ENI.MAC, toAllocate)
		if err != nil {
			for i := 0; i < toAllocate; i++ {
				resultChan <- &ENIIP{
					ENIIP: &types.ENIIP{
						ENI: eni.ENI,
					},
					err: fmt.Errorf("error assign ip for ENI: %v", err),
				}
			}
		} else {
			for _, ip := range ips {
				resultChan <- &ENIIP{
					ENIIP: &types.ENIIP{
						ENI:   eni.ENI,
						IPSet: ip,
					},
					err: nil,
				}
			}
		}
	}
}

func (eni *ENI) getIPCountLocked() int {
	return eni.pending + len(eni.ips)
}
