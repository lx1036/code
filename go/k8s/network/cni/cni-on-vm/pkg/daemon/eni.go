package daemon

import (
	"context"
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

func newENIFactory(poolConfig *types.PoolConfig, ecs ipam.API) (*eniFactory, error) {

	return &eniFactory{
		name:                   factoryNameENI,
		switches:               poolConfig.VSwitch,
		eniTags:                poolConfig.ENITags,
		securityGroup:          poolConfig.SecurityGroup,
		instanceID:             poolConfig.InstanceID,
		ecs:                    ecs,
		vswitchIPCntMap:        make(map[string]int),
		vswitchSelectionPolicy: poolConfig.VSwitchSelectionPolicy,
	}, nil
}

func (f *eniFactory) Create(int) ([]types.NetworkResource, error) {
	return f.CreateWithIPCount(1, false)
}

func (f *eniFactory) CreateWithIPCount(count int, trunk bool) ([]types.NetworkResource, error) {

	eni, err := f.ecs.AllocateENI(context.Background(), vSwitches[0], f.securityGroup, f.instanceID, trunk, count, tags)
	if err != nil {
		return nil, err
	}
	return []types.NetworkResource{eni}, nil
}
