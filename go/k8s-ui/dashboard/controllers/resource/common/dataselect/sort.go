package dataselect

import (
	"github.com/gin-gonic/gin"
	"strings"
)

type SortQuery struct {
	SortByList []SortBy
}

type SortBy struct {
	Ascending bool
	Property  PropertyName
}
type PropertyName string

func parseSortQueryFromRequest(context *gin.Context) *SortQuery {
	strings.Split(context.Query("sortBy"), ",")
}

func NewSortQuery(sortQuery []string) *SortQuery {

}
