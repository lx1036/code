package base

import (
	"github.com/astaxie/beego"
	"k8s-lx1036/k8s-ui/backend/models"
	"k8s-lx1036/k8s-ui/backend/models/response"
	"net/http"
)

type APIKeyController struct {
	beego.Controller
	
	APIKey  *models.APIKey
	Action  string
	Success response.Success
	Failure response.Failure
}


// 用于负责 get 数据的接口，当 error 列表不为空的时候，返回 error 列表
// 当 参数为 nil 的时候，返回 "200"
func (c *APIKeyController) HandleResponse(data interface{}) {
	if len(c.Failure.Body.Errors) > 0 {
		c.Failure.Body.Code = http.StatusInternalServerError
		//c.HandleByCode(http.StatusInternalServerError)
		return
	}
	if data == nil {
		c.Success.Body.Code = http.StatusOK
		data = c.Success.Body
	}
	c.publishRequestMessage(http.StatusOK, data)
	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = data
	c.ServeJSON()
}

func (c *APIKeyController) publishRequestMessage(code int, data interface{}) {

}
