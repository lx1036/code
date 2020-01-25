package permission

import (
	"k8s-lx1036/k8s-ui/backend/controllers/base"
	"k8s-lx1036/k8s-ui/backend/models"
)

type UserController struct {
	base.APIController
}

func (controller *UserController) URLMapping() {
	controller.Mapping("UserStatistics", controller.UserStatistics)
}

func (controller *UserController) Prepare() {
	controller.APIController.Prepare()
}

// @router /statistics [get]
func (controller *UserController) UserStatistics() {
	param := controller.BuildQueryParam()
	totalCount, err := models.GetTotalCount(new(models.User), param)
	if err != nil {

	}

	controller.Success(models.UserStatistics{Total: totalCount})
}
