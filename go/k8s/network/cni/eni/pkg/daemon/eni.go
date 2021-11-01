package daemon

import (
	"context"
	"sync"
	"time"

	"k8s-lx1036/k8s/network/cni/eni/pkg/ipam"
	"k8s-lx1036/k8s/network/cni/eni/pkg/pool"
	"k8s-lx1036/k8s/network/cni/eni/pkg/types"
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

type eniResourceManager struct {
	pool pool.ObjectPool
	ecs  ipam.API
}

func newENIResourceManager(poolConfig *types.PoolConfig, ecs ipam.API, allocatedResources map[string]resourceManagerInitItem,
	ipFamily *types.IPFamily) (ResourceManager, error) {

	factory, err := newENIFactory(poolConfig, ecs)
	if err != nil {
		return nil, errors.Wrapf(err, "error create ENI factory")
	}

	p, err := pool.NewSimpleObjectPool(poolCfg)
	if err != nil {
		return nil, err
	}
	mgr := &eniResourceManager{
		pool: p,
		ecs:  ecs,
	}

	return mgr, nil
}

func (manager *eniResourceManager) Allocate(context *interface{}, prefer string) (interface{}, error) {
	return manager.pool.Acquire(ctx, prefer, podInfoKey(ctx.pod.Namespace, ctx.pod.Name))
}

func (manager *eniResourceManager) Release(context *interface{}, resItem interface{}) error {
	panic("implement me")
}

func (manager *eniResourceManager) GarbageCollection(inUseResSet map[string]interface{}, expireResSet map[string]interface{}) error {
	panic("implement me")
}

func (manager *eniResourceManager) Stat(context *interface{}, resID string) (interface{}, error) {
	panic("implement me")
}
