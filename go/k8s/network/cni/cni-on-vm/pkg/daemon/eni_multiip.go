package daemon

import (
	"context"
	"k8s-lx1036/k8s/network/cni/cni-on-vm/pkg/ipam"
	"k8s-lx1036/k8s/network/cni/cni-on-vm/pkg/types"
)

type networkContext struct {
	context.Context
	resources  []types.ResourceItem
	pod        *types.PodInfo
	k8sService Kubernetes
}

type eniIPResourceManager struct {
	trunkENI *types.ENI
	pool     *SimpleObjectPool
}

func newENIIPResourceManager(poolConfig *types.PoolConfig, ecs ipam.API, k8s Kubernetes, allocatedResources map[string]resourceManagerInitItem) (*eniIPResourceManager, error) {

	poolCfg := Config{
		Name:     poolNameENIIP,
		Type:     typeNameENIIP,
		MaxIdle:  poolConfig.MaxPoolSize,
		MinIdle:  poolConfig.MinPoolSize,
		Factory:  factory,
		Capacity: capacity,
	}

	p, err := NewSimpleObjectPool(poolCfg)
	if err != nil {
		return nil, err
	}

}

func (m *eniIPResourceManager) Allocate(ctx *networkContext, id string) (types.NetworkResource, error) {
	return m.pool.Acquire(ctx, id, podInfoKey(ctx.pod.Namespace, ctx.pod.Name))
}
