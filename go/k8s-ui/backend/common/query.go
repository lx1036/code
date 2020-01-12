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
