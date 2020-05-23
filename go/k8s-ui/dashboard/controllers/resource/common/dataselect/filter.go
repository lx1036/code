package dataselect

import (
	"github.com/gin-gonic/gin"
	"strings"
)

type FilterQuery struct {
	FilterByList []FilterBy
}

type FilterBy struct {
	Property PropertyName
	Value    ComparableValue
}

func parseFilterQueryFromRequest(context *gin.Context) *FilterQuery {
	strings.Split(context.Query("filterBy"), ",")
}

func NewFilterQuery(filterQuery []string) *FilterQuery {

}
