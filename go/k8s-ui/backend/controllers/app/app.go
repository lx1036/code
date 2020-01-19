package app

import (
	"k8s-lx1036/k8s-ui/backend/controllers/base"
	"k8s-lx1036/k8s-ui/backend/models"
)

type AppController struct {
	base.APIController
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
