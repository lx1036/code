package controllers

import (
	"encoding/json"
	"k8s-lx1036/k8s-ui/backend/controllers/base"
	"k8s-lx1036/k8s-ui/backend/models"
)

type AppController struct {
	base.APIController
}

func (controller *AppController) URLMapping() {
	controller.Mapping("List", controller.List)
	controller.Mapping("AppStatistics", controller.AppStatistics)
	controller.Mapping("Update", controller.Update)
	controller.Mapping("GetNames", controller.GetNames)
	controller.Mapping("Create", controller.Create)
	controller.Mapping("Get", controller.Get)
	controller.Mapping("Delete", controller.Delete)
}

func (controller *AppController) Prepare() {
	controller.APIController.Prepare()
}

func (controller *AppController) Create() {

}

func (controller *AppController) GetNames() {

}

func (controller *AppController) AppStatistics() {
	param := controller.BuildQueryParam()
	totalCount, err := models.GetTotalCount(new(models.App), param)
	if err != nil {

	}
	details, err := models.AppModel.GetAppCountGroupByNamespace()
	if err != nil {

	}

	controller.Success(models.AppStatistics{Total: totalCount, Details: details})
}

func (controller *AppController) List() {
	param := controller.BuildQueryParam()

	starred := controller.GetBoolParamFromQueryWithDefault("starred", false)

	total, err := models.AppModel.Count(param, starred, controller.User.Id)
	if err != nil {

	}

	apps, err := models.AppModel.List(param, starred, controller.User.Id)
	if err != nil {

	}

	controller.Success(param.NewPage(total, apps))
	return
}

// @router /:id [put]
func (controller *AppController) Update() {
	id := controller.GetIdFromURL()
	var app models.App
	err := json.Unmarshal(controller.Ctx.Input.RequestBody, &app)
	if err != nil {

	}

	app.Id = id
	err = models.AppModel.UpdateById(&app)
	if err != nil {

	}

	controller.Success(app)
}
