package controllers

import (
	"fmt"
	"github.com/astaxie/beego/logs"
	"k8s-lx1036/k8s-ui/backend/controllers/base"
	"k8s-lx1036/k8s-ui/backend/models"
)

type DeploymentController struct {
	base.APIController
}

func (controller *DeploymentController) URLMapping() {
	controller.Mapping("List", controller.List)
	controller.Mapping("Get", controller.Get)
}

func (controller *DeploymentController) Prepare() {

}

// @Param name query string false "name filter"
// @router / [get]
func (controller *DeploymentController) List() {
	//params = deployment.BuildQueryParams()
	appid := controller.Ctx.Input.Param("appid")
	fmt.Println(appid)
	//controller.Success()
	return
}

// @Title Get
// @router /:id([0-9]+) [get]
func (controller *DeploymentController) Get() {
	id := controller.GetIdFromURL()
	deployment, err := models.DeploymentModel.GetById(id)
	if err != nil {
		logs.Error("get deployment by id [%d], error: %v", id, err)
		return
	}
	controller.Success(deployment)
	return
}
