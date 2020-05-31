package dataselect

import (
	"github.com/gin-gonic/gin"
	"strconv"
)

type PaginationQuery struct {
	ItemsPerPage int
	Page         int
}

func NewPaginationQuery(itemsPerPage, page int) *PaginationQuery {
	return &PaginationQuery{
		ItemsPerPage: itemsPerPage,
		Page:         page,
	}
}

func parsePaginationQueryFromRequest(context *gin.Context) *PaginationQuery {
	itemsPerPage, _ := strconv.Atoi(context.Query("itemsPerPage"))
	page, _ := strconv.Atoi(context.Query("page"))

	return NewPaginationQuery(itemsPerPage, page)
}
