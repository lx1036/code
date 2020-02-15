package controllers

import (
	"github.com/gin-gonic/gin"
	"k8s-lx1036/k8s-ui/backend/controllers/base"
	"k8s-lx1036/k8s-ui/backend/database/lorm"
	"k8s-lx1036/k8s-ui/backend/models"
	"net/http"
)

type AppController struct {
	//base.APIController
}

//func (controller *AppController) URLMapping() {
//	controller.Mapping("List", controller.List)
//	controller.Mapping("AppStatistics", controller.AppStatistics)
//	controller.Mapping("Update", controller.Update)
//	controller.Mapping("GetNames", controller.GetNames)
//	controller.Mapping("Create", controller.Create)
//	controller.Mapping("Get", controller.Get)
//	controller.Mapping("Delete", controller.Delete)
//}
//
//func (controller *AppController) Prepare() {
//	controller.APIController.Prepare()
//}

func (controller *AppController) Create() {

}

func (controller *AppController) GetNames() {

}

func (controller *AppController) AppStatistics() gin.HandlerFunc {
	return func(context *gin.Context) {
		var count int
		type Total struct {
			Total int `json:"total"`
		}
		lorm.DB.Table(models.App{}.TableName()).Count(&count)
		context.JSON(http.StatusOK, base.JsonResponse{
			Errno:  0,
			Errmsg: "success",
			Data: Total{Total:count},
		})
	}
}

func (controller *AppController) List() {
	//param := controller.BuildQueryParam()
	//
	//starred := controller.GetBoolParamFromQueryWithDefault("starred", false)
	//
	//total, err := models.AppModel.Count(param, starred, int64(controller.User.ID))
	//if err != nil {
	//
	//}
	//
	//apps, err := models.AppModel.List(param, starred, int64(controller.User.ID))
	//if err != nil {
	//
	//}
	//
	//controller.Success(param.NewPage(total, apps))
	return
}

// @router /:id [put]
func (controller *AppController) Update() {
	//id := controller.GetIdFromURL()
	//var app models.App
	//err := json.Unmarshal(controller.Ctx.Input.RequestBody, &app)
	//if err != nil {
	//
	//}
	//
	//app.ID = uint(id)
	//err = models.AppModel.UpdateById(&app)
	//if err != nil {
	//
	//}
	//
	//controller.Success(app)
}
