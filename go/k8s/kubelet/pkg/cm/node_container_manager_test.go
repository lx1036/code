package cm

import (
	"github.com/stretchr/testify/assert"
	"testing"

	evictionapi "k8s-lx1036/k8s/kubelet/pkg/eviction/api"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

// INFO: 测试 nodeReserved=kubeReserved + systemReserved + evictionReserved
func TestNodeAllocatableReservationForScheduling(test *testing.T) {
	memoryEvictionThreshold := resource.MustParse("100Mi")
	cpuMemCases := []struct {
		description    string
		kubeReserved   v1.ResourceList
		systemReserved v1.ResourceList
		expected       v1.ResourceList
		capacity       v1.ResourceList
		hardThreshold  evictionapi.ThresholdValue
	}{
		// INFO: nodeReserved = kubeReserved + systemReserved + evictionReserved
		{
			// cpu: 150m = 100m + 50m, memory: 150Mi = 100Mi + 50Mi
			description:    "kubeReserved-systemReserved",
			kubeReserved:   getResourceList("100m", "100Mi"),
			systemReserved: getResourceList("50m", "50Mi"),
			capacity:       getResourceList("10", "10Gi"),
			expected:       getResourceList("150m", "150Mi"),
		},
		{
			// cpu: 150m = 100m + 50m, memory: 250Mi = 100Mi + 50Mi + memoryEvictionThreshold(100Mi)
			description:    "kubeReserved-systemReserved-evictionReserved",
			kubeReserved:   getResourceList("100m", "100Mi"),
			systemReserved: getResourceList("50m", "50Mi"),
			hardThreshold: evictionapi.ThresholdValue{
				Quantity: &memoryEvictionThreshold,
			},
			capacity: getResourceList("10", "10Gi"),
			expected: getResourceList("150m", "250Mi"),
		},
		{
			// cpu: 150m = 100m + 50m, memory: 250Mi = 100Mi + 50Mi + 10Gi * 0.05
			description:    "kubeReserved-systemReserved-evictionReserved",
			kubeReserved:   getResourceList("100m", "100Mi"),
			systemReserved: getResourceList("50m", "50Mi"),
			capacity:       getResourceList("10", "10Gi"),
			hardThreshold: evictionapi.ThresholdValue{
				Percentage: 0.05,
			},
			expected: getResourceList("150m", "694157320"),
		},
		{
			description:    "no-reserved",
			kubeReserved:   v1.ResourceList{},
			systemReserved: v1.ResourceList{},
			capacity:       getResourceList("10", "10Gi"),
			expected:       getResourceList("", ""),
		},
		{
			// cpu: 50m = 0 + 50m, memory: 150Mi = 100Mi + 50Mi
			description:    "kubeReserved-systemReserved",
			kubeReserved:   getResourceList("", "100Mi"),
			systemReserved: getResourceList("50m", "50Mi"),
			capacity:       getResourceList("10", "10Gi"),
			expected:       getResourceList("50m", "150Mi"),
		},
		{
			// cpu: 50m = 50m + 0, memory: 150Mi = 100Mi + 50Mi
			description:    "kubeReserved-systemReserved",
			kubeReserved:   getResourceList("50m", "100Mi"),
			systemReserved: getResourceList("", "50Mi"),
			capacity:       getResourceList("10", "10Gi"),
			expected:       getResourceList("50m", "150Mi"),
		},
		{
			// memory: 150Mi = 100Mi + 50Mi
			description:    "kubeReserved-systemReserved",
			kubeReserved:   getResourceList("", "100Mi"),
			systemReserved: getResourceList("", "50Mi"),
			capacity:       getResourceList("10", ""),
			expected:       getResourceList("", "150Mi"),
		},
	}

	// INFO: 测试 nodeReserved
	for _, cpuMemCase := range cpuMemCases {
		test.Run(cpuMemCase.description, func(t *testing.T) {
			nodeConfig := NodeConfig{
				NodeAllocatableConfig: NodeAllocatableConfig{
					KubeReserved:   cpuMemCase.kubeReserved,
					SystemReserved: cpuMemCase.systemReserved,
					HardEvictionThresholds: []evictionapi.Threshold{
						{
							Signal:   evictionapi.SignalMemoryAvailable,
							Operator: evictionapi.OpLessThan,
							Value:    cpuMemCase.hardThreshold,
						},
					},
				},
			}
			containerManager := &containerManagerImpl{
				NodeConfig: nodeConfig,
				capacity:   cpuMemCase.capacity,
			}

			for resourceName, quantity := range containerManager.GetNodeAllocatableReservation() {
				expected, exists := cpuMemCase.expected[resourceName]
				assert.True(t, exists, "test case expected resource %q", resourceName)
				assert.Equal(t, expected.MilliValue(), quantity.MilliValue(), "test case failed for resource %q", resourceName)
			}
		})
	}

	// INFO: 测试 ephemeral storage
	ephemeralStorageTestCases := []struct {
		description   string
		kubeReserved  v1.ResourceList
		expected      v1.ResourceList
		capacity      v1.ResourceList
		hardThreshold evictionapi.ThresholdValue
	}{
		{
			// storage: 100Mi = 100Mi + 0
			description:  "kubeReserved ephemeral storage",
			kubeReserved: getEphemeralStorageResourceList("100Mi"),
			capacity:     getEphemeralStorageResourceList("10Gi"),
			expected:     getEphemeralStorageResourceList("100Mi"),
		},
	}

	for _, ephemeralStorageTestCase := range ephemeralStorageTestCases {
		test.Run(ephemeralStorageTestCase.description, func(t *testing.T) {
			nodeConfig := NodeConfig{
				NodeAllocatableConfig: NodeAllocatableConfig{
					KubeReserved: ephemeralStorageTestCase.kubeReserved,
					HardEvictionThresholds: []evictionapi.Threshold{
						{
							Signal:   evictionapi.SignalNodeFsAvailable,
							Operator: evictionapi.OpLessThan,
							Value:    ephemeralStorageTestCase.hardThreshold,
						},
					},
				},
			}
			containerManager := &containerManagerImpl{
				NodeConfig: nodeConfig,
				capacity:   ephemeralStorageTestCase.capacity,
			}

			for resourceName, quantity := range containerManager.GetNodeAllocatableReservation() {
				expected, exists := ephemeralStorageTestCase.expected[resourceName]
				assert.True(t, exists, "test case expected resource %q", resourceName)
				assert.Equal(t, expected.MilliValue(), quantity.MilliValue(), "test case failed for resource %q", resourceName)
			}
		})
	}
}

func TestNodeAllocatableAbsolute(test *testing.T) {
	memoryEvictionThreshold := resource.MustParse("100Mi")
	cpuMemCases := []struct {
		description    string
		kubeReserved   v1.ResourceList
		systemReserved v1.ResourceList
		capacity       v1.ResourceList
		expected       v1.ResourceList
		hardThreshold  evictionapi.ThresholdValue
	}{
		{
			description:    "10-100m-50m",
			kubeReserved:   getResourceList("100m", "100Mi"),
			systemReserved: getResourceList("50m", "50Mi"),
			capacity:       getResourceList("10", "10Gi"),
			expected:       getResourceList("9850m", "10090Mi"),
		},
		{
			description:    "10-100m-50m,10Gi-100Mi-50Mi-100Mi",
			kubeReserved:   getResourceList("100m", "100Mi"),
			systemReserved: getResourceList("50m", "50Mi"),
			hardThreshold: evictionapi.ThresholdValue{
				Quantity: &memoryEvictionThreshold,
			},
			capacity: getResourceList("10", "10Gi"),
			expected: getResourceList("9850m", "10090Mi"),
		},
		{
			description:    "10-100m-50m,10Gi-100Mi-50Mi-10Gi*0.05",
			kubeReserved:   getResourceList("100m", "100Mi"),
			systemReserved: getResourceList("50m", "50Mi"),
			hardThreshold: evictionapi.ThresholdValue{
				Percentage: 0.05,
			},
			capacity: getResourceList("10", "10Gi"),
			expected: getResourceList("9850m", "10090Mi"),
		},
		{
			description:    "10,10Gi",
			kubeReserved:   v1.ResourceList{},
			systemReserved: v1.ResourceList{},
			capacity:       getResourceList("10", "10Gi"),
			expected:       getResourceList("10", "10Gi"),
		},
		{
			description:    "10-50m,10Gi-100Mi-50Mi",
			kubeReserved:   getResourceList("", "100Mi"),
			systemReserved: getResourceList("50m", "50Mi"),
			capacity:       getResourceList("10", "10Gi"),
			expected:       getResourceList("9950m", "10090Mi"),
		},

		{
			description:    "10-50m,10Gi-100Mi-50Mi",
			kubeReserved:   getResourceList("50m", "100Mi"),
			systemReserved: getResourceList("", "50Mi"),
			capacity:       getResourceList("10", "10Gi"),
			expected:       getResourceList("9950m", "10090Mi"),
		},
		{
			description:    "10,0-100Mi-50Mi",
			kubeReserved:   getResourceList("", "100Mi"),
			systemReserved: getResourceList("", "50Mi"),
			capacity:       getResourceList("10", ""),
			expected:       getResourceList("10", ""),
		},
	}

	// INFO: 测试 allocatable 可分配资源量
	for _, cpuMemCase := range cpuMemCases {
		test.Run(cpuMemCase.description, func(t *testing.T) {
			nodeConfig := NodeConfig{
				NodeAllocatableConfig: NodeAllocatableConfig{
					KubeReserved:   cpuMemCase.kubeReserved,
					SystemReserved: cpuMemCase.systemReserved,
					HardEvictionThresholds: []evictionapi.Threshold{
						{
							Signal:   evictionapi.SignalMemoryAvailable,
							Operator: evictionapi.OpLessThan,
							Value:    cpuMemCase.hardThreshold,
						},
					},
				},
			}
			cm := &containerManagerImpl{
				NodeConfig: nodeConfig,
				capacity:   cpuMemCase.capacity,
			}
			for resourceName, quantity := range cm.getNodeAllocatableAbsolute() {
				expected, exists := cpuMemCase.expected[resourceName]
				assert.True(t, exists)
				assert.Equal(t, expected.MilliValue(), quantity.MilliValue(), "test case failed for resource %q", resourceName)
			}
		})
	}
}

func getEphemeralStorageResourceList(storage string) v1.ResourceList {
	res := v1.ResourceList{}
	if storage != "" {
		res[v1.ResourceEphemeralStorage] = resource.MustParse(storage)
	}
	return res
}
