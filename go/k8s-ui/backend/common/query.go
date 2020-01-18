package common

type QueryParam struct {
	PageNo   int64                  `json:"pageNo"`
	PageSize int64                  `json:"pageSize"`
	Query    map[string]interface{} `json:"query"`
	Sortby   string                 `json:"sortby"`
	Groupby  []string               `json:"groupby"`
	Relate   string                 `json:"relate"`
	// only for kubernetes resource
	LabelSelector string `json:"-"`
}

type Page struct {
	PageNo     int64       `json:"pageNo"`
	PageSize   int64       `json:"pageSize"`
	TotalPage  int64       `json:"totalPage"`
	TotalCount int64       `json:"totalCount"`
	List       interface{} `json:"list"`
}

func (param *QueryParam) Limit() int64 {
	return param.PageSize
}

func (param *QueryParam) Offset() interface{} {
	offset := (param.PageNo - 1) * param.PageSize
	if offset < 0 {
		offset = 0
	}

	return offset
}

func (param *QueryParam) NewPage(total int64, list interface{}) *Page {
	count := total / param.PageSize
	if total%param.PageSize > 0 {
		count += 1
	}

	return &Page{
		PageNo:     param.PageNo,
		PageSize:   param.PageSize,
		TotalPage:  count,
		TotalCount: total,
		List:       list,
	}
}
