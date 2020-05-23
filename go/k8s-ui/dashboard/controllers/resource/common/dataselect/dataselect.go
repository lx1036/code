package dataselect

import "github.com/gin-gonic/gin"

type DataSelectQuery struct {
	PaginationQuery *PaginationQuery
	SortQuery *SortQuery
	FilterQuery *FilterQuery
	MetricQuery *MetricQuery
}

type PaginationQuery struct {
	ItemsPerPage int
	Page int
}

type SortQuery struct {
	SortByList []SortBy
}
type SortBy struct {
	Ascending bool
	Property PropertyName
}
type PropertyName string

func ParseDataSelectFromRequest(context *gin.Context)  {
	
}

