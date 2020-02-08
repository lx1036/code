package controllers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"k8s-lx1036/k8s-ui/backend/controllers/base"
	"k8s-lx1036/k8s-ui/backend/database/lorm"
	"k8s-lx1036/k8s-ui/backend/models"
	"k8s-lx1036/k8s-ui/backend/routers-gin/middlewares"
	"net/http"
)

type NotificationController struct {
}

func (controller *NotificationController) URLMapping() {
	//	controller.Mapping("List", controller.List)
	//	controller.Mapping("Create", controller.Create)
	//	controller.Mapping("Publish", controller.Publish)
	//	controller.Mapping("Subscribe", controller.Subscribe)
	//	controller.Mapping("Read", controller.Read)
}

func (controller *NotificationController) Prepare() {
	//controller.APIController.Prepare()
	//_, actionName := controller.GetControllerAndAction()
	//switch actionName {
	//case "List", "Create", "Publish":
	//	if !controller.User.Admin {
	//		// 只有管理员才能查看、创建和广播消息
	//		controller.AbortForbidden("only admin user can list/create/publish notification.")
	//	}
	//default:
	//}
}

// @router / [get]
func (controller *NotificationController) List() {
	//param := controller.BuildQueryParam()
	//
	//totalCount, err := models.GetTotalCount(new(models.Notification), param)
	//if err != nil {
	//
	//}
	//var notifications []models.Notification
	//err = models.GetAll(new(models.Notification), &notifications, param)
	//if err != nil {
	//
	//}
	//
	//controller.Success(param.NewPage(totalCount, notifications))
}

func (controller *NotificationController) Create() {

}

func (controller *NotificationController) Publish() {

}

func (controller *NotificationController) Subscribe() gin.HandlerFunc {
	return func(context *gin.Context) {
		user := middlewares.User
		db := lorm.DB.Debug().Where("user_id=?", user.ID)
		isRead := context.Query("is_read")
		if isRead != "" {
			db.Where("is_read=?", isRead)
		}

		var notificationLogs []models.NotificationLog
		err := db.Limit(100).Find(&notificationLogs).Error
		if err != nil {
			context.JSON(http.StatusNoContent, base.JsonResponse{
				Errno:  -1,
				Errmsg: fmt.Sprintf("failed: empty content [%s]", err.Error()),
				Data:   nil,
			})
			return
		}

		context.JSON(http.StatusOK, base.JsonResponse{
			Errno:  0,
			Errmsg: "success",
			Data:   notificationLogs,
		})
	}
}

func (controller *NotificationController) Read() {

}
