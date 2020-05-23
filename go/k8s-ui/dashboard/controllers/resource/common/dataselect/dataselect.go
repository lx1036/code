package dataselect

import (
	"github.com/gin-gonic/gin"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
)


var ListEverything = metav1.ListOptions{
	LabelSelector: labels.Everything().String(),
	FieldSelector: fields.Everything().String(),
}


type DataSelectQuery struct {
	PaginationQuery *PaginationQuery
	SortQuery *SortQuery
	FilterQuery *FilterQuery
	MetricQuery *MetricQuery
}



// ComparableValue hold any value that can be compared to its own kind.
type ComparableValue interface {
	// Compares self with other value. Returns 1 if other value is smaller, 0 if they are the same, -1 if other is larger.
	Compare(ComparableValue) int
	// Returns true if self value contains or is equal to other value, false otherwise.
	Contains(ComparableValue) bool
}

func NewDataSelectQuery(paginationQuery *PaginationQuery, sortQuery *SortQuery, filterQuery *FilterQuery, metricQuery *MetricQuery) *DataSelectQuery {
	return &DataSelectQuery{
		PaginationQuery: paginationQuery,
		SortQuery: sortQuery,
		FilterQuery: filterQuery,
		MetricQuery: metricQuery,
	}
}

func ParseDataSelectFromRequest(context *gin.Context) *DataSelectQuery {
	paginationQuery := parsePaginationQueryFromRequest(context)
	sortQuery := parseSortQueryFromRequest(context)
	filterQuery := parseFilterQueryFromRequest(context)
	metricQuery := parseMetricQueryFromRequest(context)

	return &DataSelectQuery{
		PaginationQuery: paginationQuery,
		SortQuery: sortQuery,
		FilterQuery: filterQuery,
		MetricQuery: metricQuery,
	}
}


