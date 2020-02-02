package controllers

import (
	"github.com/astaxie/beego/logs"
	"k8s-lx1036/k8s-ui/backend/controllers/base"
	"k8s-lx1036/k8s-ui/backend/models"
)

type CronjobController struct {
	base.APIController
}

func (controller *CronjobController) URLMapping() {
	controller.Mapping("GetNames", controller.GetNames)
	controller.Mapping("List", controller.List)
	controller.Mapping("Create", controller.Create)
	controller.Mapping("Get", controller.Get)
	controller.Mapping("Update", controller.Update)
	controller.Mapping("Delete", controller.Delete)
}

func (controller *CronjobController) Prepare() {
	// Check administration
	controller.APIController.Prepare()

	// Check permission
	perAction := ""
	_, method := controller.GetControllerAndAction()
	switch method {
	case "Get", "List", "GetNames":
		perAction = models.PermissionRead
	case "Create":
		perAction = models.PermissionCreate
	case "Update":
		perAction = models.PermissionUpdate
	case "Delete":
		perAction = models.PermissionDelete
	}

	if perAction != "" {
		controller.CheckPermission(models.PermissionTypeCronjob, perAction)
	}
}

func (controller *CronjobController) List() {

}

func (controller *CronjobController) Create() {

}
func (controller *CronjobController) GetNames() {

}

func (controller *CronjobController) Update() {

}

func (controller *CronjobController) Get() {
	id := controller.GetIdFromURL()
	cronjob, err := models.CronjobModel.GetById(int64(id))
	if err != nil {
		logs.Error("get by id (%d) error.%v", id, err)
		controller.HandleError(err)
		return
	}

	controller.Success(cronjob)
	return
}

func (controller *CronjobController) HandleError(e error) {

}
