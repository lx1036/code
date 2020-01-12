package models

import (
	"github.com/astaxie/beego/orm"
	"k8s-lx1036/k8s-ui/backend/common"
	"strings"
)

func GetAll(queryTable interface{}, list interface{}, query *common.QueryParam) error {
	qs := Ormer().QueryTable(queryTable)
	qs = BuildFilter(qs, query.Query)
	if query.Relate != "" {

	}

	if len(query.Groupby) != 0 {
		qs = qs.GroupBy(query.Groupby...)
	}
	if query.Sortby != "" {
		qs = qs.OrderBy(query.Sortby)
	}

	if _, err := qs.Limit(query.Limit(), query.Offset()).All(list); err != nil {
		return err
	}

	return nil
}

func BuildFilter(querySeter orm.QuerySeter, query map[string]interface{}) orm.QuerySeter {
	for key, value := range query {
		key = strings.Replace(key, ".", ListFilterExprSep, -1)
		querySeter = querySeter.Filter(key, value)
	}

	return querySeter
}
