package metrics

import "fmt"

func GetResourceUtilizationRatio(metrics PodMetricsInfo,
	requests map[string]int64, targetUtilization int32) (utilizationRatio float64,
	currentUtilization int32, rawAverageValue int64, err error) {
	metricsTotal := int64(0)
	requestsTotal := int64(0)
	numEntries := 0

	for podName, metric := range metrics {
		request, hasRequest := requests[podName]
		if !hasRequest {
			// we check for missing requests elsewhere, so assuming missing requests == extraneous metrics
			continue
		}

		metricsTotal += metric.Value
		requestsTotal += request
		numEntries++
	}

	// if the set of requests is completely disjoint from the set of metrics,
	// then we could have an issue where the requests total is zero
	if requestsTotal == 0 {
		return 0, 0, 0, fmt.Errorf("no metrics returned matched known pods")
	}
	// 当前实际利用率
	currentUtilization = int32((metricsTotal * 100) / requestsTotal)

	// 计算ratio比率 float64(currentUtilization) / float64(targetUtilization)
	return float64(currentUtilization) / float64(targetUtilization), currentUtilization, metricsTotal / int64(numEntries), nil
}
