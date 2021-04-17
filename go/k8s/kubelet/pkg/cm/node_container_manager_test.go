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

}
