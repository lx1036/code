package collector

import (
	"k8s-lx1036/k8s/meetup/2020-12-31/u26/request"
	"k8s-lx1036/k8s/meetup/2020-12-31/u26/storage"
)

type MetricsCollector struct {
	storage storage.Storage
}

func (collector *MetricsCollector) RecordRequest(requestInfo request.RequestInfo) {
	if len(requestInfo.ApiName) == 0 {
		return
	}

	collector.storage.SaveRequestInfo(requestInfo)
}

func NewMetricsCollector(storage storage.Storage) *MetricsCollector {
	return &MetricsCollector{
		storage: storage,
	}
}
