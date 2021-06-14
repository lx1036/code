package prometheus

import (
	prommodel "github.com/prometheus/common/model"
	"k8s-lx1036/k8s/monitor/vpa/recommender/pkg/controller/clusterstate/types"
)

func promMetricToLabelMap(metric prommodel.Metric) map[string]string {
	labels := map[string]string{}
	for k, v := range metric {
		labels[string(k)] = string(v)
	}
	return labels
}

func getContainerUsageSamplesFromSamples(samples []prommodel.SamplePair, resource types.ResourceName) []types.ContainerUsageSample {
	res := make([]types.ContainerUsageSample, 0)
	for _, sample := range samples {
		res = append(res, types.ContainerUsageSample{
			MeasureStart: sample.Timestamp.Time(),
			Usage:        resourceAmountFromValue(float64(sample.Value), resource),
			Resource:     resource,
		})
	}
	return res
}

func resourceAmountFromValue(value float64, resource types.ResourceName) types.ResourceAmount {
	// This assumes CPU value is in cores and memory in bytes, which is true
	// for the metrics this class queries from Prometheus.
	switch resource {
	case types.ResourceCPU:
		return CPUAmountFromCores(value)
	case types.ResourceMemory:
		return MemoryAmountFromBytes(value)
	}
	return types.ResourceAmount(0)
}

// CPUAmountFromCores converts CPU cores to a ResourceAmount.
func CPUAmountFromCores(cores float64) types.ResourceAmount {
	return resourceAmountFromFloat(cores * 1000.0)
}

func resourceAmountFromFloat(amount float64) types.ResourceAmount {
	if amount < 0 {
		return types.ResourceAmount(0)
	} else if amount > float64(types.MaxResourceAmount) {
		return types.MaxResourceAmount
	} else {
		return types.ResourceAmount(amount)
	}
}

// MemoryAmountFromBytes converts memory bytes to a ResourceAmount.
func MemoryAmountFromBytes(bytes float64) types.ResourceAmount {
	return resourceAmountFromFloat(bytes)
}
