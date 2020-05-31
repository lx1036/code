package dataselect

import (
	"github.com/gin-gonic/gin"
	"strings"
)

// MetricQuery holds parameters for metric extraction process.
// It accepts list of metrics to be downloaded and a list of aggregations that should be performed for each metric.
// Query has this format  metrics=metric1,metric2,...&aggregations=aggregation1,aggregation2,...
type MetricQuery struct {
	// Metrics to download, all available metric names can be found here:
	// https://github.com/kubernetes/heapster/blob/master/docs/storage-schema.md
	MetricNames []string
	// Aggregations to be performed for each metric. Check available aggregations in aggregation.go.
	// If empty, default aggregation will be used (sum).
	Aggregations AggregationModes
}

type AggregationModes []AggregationMode

// AggregationMode informs how data should be aggregated (sum, min, max)
type AggregationMode string

func parseMetricQueryFromRequest(context *gin.Context) *MetricQuery {
	metricNames := strings.Split(context.Query("metricNames"), ",")
	aggregations := strings.Split(context.Query("aggregations"), ",")
	var aggregationModes AggregationModes
	for _, aggregation := range aggregations {
		aggregationModes = append(aggregationModes, AggregationMode(aggregation))
	}

	return NewMetricQuery(metricNames, aggregationModes)
}

func NewMetricQuery(metricNames []string, aggregations AggregationModes) *MetricQuery {
	return &MetricQuery{
		MetricNames:  metricNames,
		Aggregations: aggregations,
	}
}
