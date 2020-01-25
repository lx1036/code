package base

import (
	"k8s-lx1036/k8s-ui/backend/common"
	"strconv"
)

type ParamBuilderController struct {
	ResultHandlerController
}

func (controller *ParamBuilderController) GetIdFromURL() int64 {
	return controller.GetIntParamFromURL(":id")
}

func (controller *ParamBuilderController) GetIntParamFromURL(param string) int64 {
	paramStr := controller.Ctx.Input.Param(param)
	if len(paramStr) == 0 {
		//c.AbortBadRequest(fmt.Sprintf("Invalid %s in URL", param))
	}

	paramInt, err := strconv.ParseInt(paramStr, 10, 64)
	if err != nil || paramInt < 0 {
		//c.AbortBadRequest(fmt.Sprintf("Invalid %s in URL", param))
	}

	return paramInt
}

func (controller *ParamBuilderController) BuildQueryParam() *common.QueryParam {
	no, size := controller.buildPageParam()

	qmap := map[string]interface{}{}
	deletedStr := controller.Input().Get("deleted")
	if deletedStr != "" {

	}

	filter := controller.Input().Get("filter")
	if filter != "" {

	}

	relate := ""
	if controller.Input().Get("relate") != "" {
		relate = controller.Input().Get("relate")
	}

	return &common.QueryParam{
		PageNo:   no,
		PageSize: size,
		Query:    qmap,
		Sortby:   controller.Input().Get("sortby"),
		Relate:   relate,
	}
}

const (
	defaultPageNo   = 1
	defaultPageSize = 10
)

func (controller *ParamBuilderController) buildPageParam() (no int64, size int64) {
	pageNo := controller.Input().Get("pageNo")
	pageSize := controller.Input().Get("pageSize")
	if pageNo == "" {
		pageNo = strconv.Itoa(defaultPageNo)
	}
	if pageSize == "" {
		pageSize = strconv.Itoa(defaultPageSize)
	}
	no, err := strconv.ParseInt(pageNo, 10, 64)
	if err != nil || no < 1 {
		controller.AbortBadRequest("Invalid pageNo in query.")
	}
	size, err = strconv.ParseInt(pageSize, 10, 64)
	if err != nil || size < 1 {
		controller.AbortBadRequest("Invalid pageSize in query.")
	}
	return
}
