package base

import (
	"encoding/json"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"k8s-lx1036/wayne/backend/models/response/errors"
	"k8s-lx1036/wayne/backend/util/hack"
	"net/http"
)

type ResultHandlerController struct {
	beego.Controller
}

// Abort stops controller handler and show the error data， e.g. Prepare
func (c *ResultHandlerController) AbortForbidden(msg string) {
	logs.Info("Abort Forbidden error. %s", msg)
	c.CustomAbort(http.StatusForbidden, hack.String(c.errorResult(http.StatusForbidden, msg)))
}

func (c *ResultHandlerController) errorResult(code int, msg string) []byte {
	errorResult := errors.ErrorResult{
		Code: code,
		Msg:  msg,
	}
	body, err := json.Marshal(errorResult)
	if err != nil {
		logs.Error("Json Marshal error. %v", err)
		c.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}
	return body
}
