package prometheus

import (
	prommodel "github.com/prometheus/common/model"
	"time"
)

type ResourceName string

type ResourceAmount int64

// Resources is a map from resource name to the corresponding ResourceAmount.
type Resources map[ResourceName]ResourceAmount

const (
	// ResourceCPU represents CPU in millicores (1core = 1000millicores).
	ResourceCPU ResourceName = "cpu"
	// ResourceMemory represents memory, in bytes. (500Gi = 500GiB = 500 * 1024 * 1024 * 1024).
	ResourceMemory ResourceName = "memory"
	// MaxResourceAmount is the maximum allowed value of resource amount.
	MaxResourceAmount = ResourceAmount(1e14)
)

type ContainerUsageSample struct {
	// Start of the measurement interval.
	MeasureStart time.Time
	// Average CPU usage in cores or memory usage in bytes.
	Usage ResourceAmount
	// CPU or memory request at the time of measurment.
	Request ResourceAmount
	// Which resource is this sample for.
	Resource ResourceName
}

type ContainerID struct {
	PodID
	// ContainerName is the name of the container, unique within a pod.
	ContainerName string
}

type PodID struct {
	Namespace string
	PodName   string
}

func promMetricToLabelMap(metric prommodel.Metric) map[string]string {
	labels := map[string]string{}
	for k, v := range metric {
		labels[string(k)] = string(v)
	}
	return labels
}

func getContainerUsageSamplesFromSamples(samples []prommodel.SamplePair, resource ResourceName) []ContainerUsageSample {
	res := make([]ContainerUsageSample, 0)
	for _, sample := range samples {
		res = append(res, ContainerUsageSample{
			MeasureStart: sample.Timestamp.Time(),
			Usage:        resourceAmountFromValue(float64(sample.Value), resource),
			Resource:     resource,
		})
	}
	return res
}

func resourceAmountFromValue(value float64, resource ResourceName) ResourceAmount {
	// This assumes CPU value is in cores and memory in bytes, which is true
	// for the metrics this class queries from Prometheus.
	switch resource {
	case ResourceCPU:
		return CPUAmountFromCores(value)
	case ResourceMemory:
		return MemoryAmountFromBytes(value)
	}
	return ResourceAmount(0)
}

// CPUAmountFromCores converts CPU cores to a ResourceAmount.
func CPUAmountFromCores(cores float64) ResourceAmount {
	return resourceAmountFromFloat(cores * 1000.0)
}

func resourceAmountFromFloat(amount float64) ResourceAmount {
	if amount < 0 {
		return ResourceAmount(0)
	} else if amount > float64(MaxResourceAmount) {
		return MaxResourceAmount
	} else {
		return ResourceAmount(amount)
	}
}

// MemoryAmountFromBytes converts memory bytes to a ResourceAmount.
func MemoryAmountFromBytes(bytes float64) ResourceAmount {
	return resourceAmountFromFloat(bytes)
}
