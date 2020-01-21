package base

import (
	"encoding/json"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"k8s-lx1036/k8s-ui/backend/models/response/errors"
	"k8s-lx1036/k8s-ui/backend/util/hack"
	"net/http"
)

type Result struct {
	Data interface{} `json:"data"`
}

type ResultHandlerController struct {
	beego.Controller
}

// Abort stops controller handler and show the error data， e.g. Prepare
func (controller *ResultHandlerController) AbortForbidden(msg string) {
	logs.Info("Abort Forbidden error. %s", msg)
	controller.CustomAbort(http.StatusForbidden, hack.String(controller.errorResult(http.StatusForbidden, msg)))
}

func (controller *ResultHandlerController) errorResult(code int, msg string) []byte {
	errorResult := errors.ErrorResult{
		Code: code,
		Msg:  msg,
	}
	body, err := json.Marshal(errorResult)
	if err != nil {
		logs.Error("Json Marshal error. %v", err)
		controller.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}
	return body
}

func (controller *ResultHandlerController) Success(data interface{}) {
	controller.Ctx.Output.SetStatus(http.StatusOK)
	controller.Data["json"] = Result{Data: data}
	controller.ServeJSON()
}

func (controller *ResultHandlerController) AbortBadRequest(msg string) {
	logs.Info("Abort BadRequest error. %s", msg)
	controller.CustomAbort(http.StatusBadRequest, hack.String(controller.errorResult(http.StatusBadRequest, msg)))
}
