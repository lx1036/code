package controllers

import (
	"github.com/gin-gonic/gin"
	"k8s-lx1036/k8s-ui/backend/controllers/base"
	"k8s-lx1036/k8s-ui/backend/database/lorm"
	"k8s-lx1036/k8s-ui/backend/models"
	"net/http"
)

type UserController struct {
}

//func (controller *UserController) URLMapping() {
//	controller.Mapping("UserStatistics", controller.UserStatistics)
//}
//
//func (controller *UserController) Prepare() {
//	controller.APIController.Prepare()
//}

// @router /statistics [get]
func (controller *UserController) UserStatistics() gin.HandlerFunc {
	return func(context *gin.Context) {
		var count int
		type Total struct {
			Total int `json:"total"`
		}
		lorm.DB.Table(models.User{}.TableName()).Count(&count)
		context.JSON(http.StatusOK, base.JsonResponse{
			Errno:  0,
			Errmsg: "success",
			Data:   Total{Total: count},
		})
	}

}
