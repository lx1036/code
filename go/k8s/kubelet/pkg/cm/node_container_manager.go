package cm

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

// INFO: 对于node层面的资源，kubernetes会将一个node上面的资源按照使用对象分为三部分：
// 1. 业务进程使用的资源， 即pods使用的资源；
// 2. kubernetes组件使用的资源，例如kubelet, docker；
// 3. 系统组件使用的资源，例如logind, journald等进程；

// GetNodeAllocatableReservation 获取 node reserved 资源:
// allocatable = capacity - kubeReserved - systemReserved - evictionReserved
// nodeReserved = kubeReserved + systemReserved + evictionReserved
func (containerManager *containerManagerImpl) GetNodeAllocatableReservation() v1.ResourceList {
	result := make(v1.ResourceList)
	evictionReservation := hardEvictionReservation(containerManager.HardEvictionThresholds, containerManager.capacity)
	for resourceName := range containerManager.capacity {
		value := resource.NewQuantity(0, resource.DecimalSI)
		if containerManager.NodeConfig.SystemReserved != nil {
			value.Add(containerManager.NodeConfig.SystemReserved[resourceName])
		}
		if containerManager.NodeConfig.KubeReserved != nil {
			value.Add(containerManager.NodeConfig.KubeReserved[resourceName])
		}
		if evictionReservation != nil {
			value.Add(evictionReservation[resourceName])
		}

		if !value.IsZero() {
			result[resourceName] = *value
		}
	}

	return result
}

// getNodeAllocatableAbsolute allocatable = capacity - systemReserved - kubeReserved 获取可分配资源量
func (containerManager *containerManagerImpl) getNodeAllocatableAbsolute() v1.ResourceList {
	return containerManager.getNodeAllocatableAbsoluteImpl(containerManager.capacity)
}
func (containerManager *containerManagerImpl) getNodeAllocatableAbsoluteImpl(capacity v1.ResourceList) v1.ResourceList {
	// INFO: allocatable = capacity - systemReserved - kubeReserved
	result := make(v1.ResourceList)
	for resourceName, quantity := range capacity {
		value := quantity.DeepCopy()
		if containerManager.NodeConfig.SystemReserved != nil {
			value.Sub(containerManager.NodeConfig.SystemReserved[resourceName])
		}
		if containerManager.NodeConfig.KubeReserved != nil {
			value.Sub(containerManager.NodeConfig.KubeReserved[resourceName])
		}
		if value.Sign() < 0 {
			// Negative Allocatable resources don't make sense.
			value.Set(0)
		}
		result[resourceName] = value
	}

	return result
}
