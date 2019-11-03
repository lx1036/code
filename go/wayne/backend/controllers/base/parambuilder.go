package base

import (
	"fmt"
	"strconv"
)

type ParamBuilderController struct {
	ResultHandlerController
}

func (c *ParamBuilderController) GetIDFromURL() int64 {
	return c.GetIntParamFromURL(":id")
}

func (c *ParamBuilderController) GetIntParamFromURL(param string) int64 {
	paramStr := c.Ctx.Input.Param(param)
	if len(paramStr) == 0 {
		c.AbortBadRequest(fmt.Sprintf("Invalid %s in URL", param))
	}

	paramInt, err := strconv.ParseInt(paramStr, 10, 64)
	if err != nil || paramInt < 0 {
		c.AbortBadRequest(fmt.Sprintf("Invalid %s in URL", param))
	}

	return paramInt
}









