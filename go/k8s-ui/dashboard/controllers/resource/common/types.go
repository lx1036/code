package common

import "time"

type ResourceStatus struct {
	// Number of resources that are currently in running state.
	Running int `json:"running"`
	
	// Number of resources that are currently in pending state.
	Pending int `json:"pending"`
	
	// Number of resources that are in failed state.
	Failed int `json:"failed"`
	
	// Number of resources that are in succeeded state.
	Succeeded int `json:"succeeded"`
}

type MetricPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     uint64    `json:"value"`
}

type Metric struct {
	MetricPoints []MetricPoint `json:"metricPoints"`
	MetricName string `json:"metricName"`
}
