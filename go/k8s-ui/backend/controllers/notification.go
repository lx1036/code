package controllers

import (
	"k8s-lx1036/k8s-ui/backend/controllers/base"
	"k8s-lx1036/k8s-ui/backend/models"
)

type NotificationController struct {
	base.APIController
}

func (controller *NotificationController) URLMapping() {
	controller.Mapping("List", controller.List)
	controller.Mapping("Create", controller.Create)
	controller.Mapping("Publish", controller.Publish)
	controller.Mapping("Subscribe", controller.Subscribe)
	controller.Mapping("Read", controller.Read)
}

func (controller *NotificationController) Prepare() {
	controller.APIController.Prepare()
	_, actionName := controller.GetControllerAndAction()
	switch actionName {
	case "List", "Create", "Publish":
		if !controller.User.Admin {
			// 只有管理员才能查看、创建和广播消息
			controller.AbortForbidden("only admin user can list/create/publish notification.")
		}
	default:
	}
}

// @router / [get]
func (controller *NotificationController) List() {
	param := controller.BuildQueryParam()

	totalCount, err := models.GetTotalCount(new(models.Notification), param)
	if err != nil {

	}
	var notifications []models.Notification
	err = models.GetAll(new(models.Notification), &notifications, param)
	if err != nil {

	}

	controller.Success(param.NewPage(totalCount, notifications))
}

func (controller *NotificationController) Create() {

}

func (controller *NotificationController) Publish() {

}

func (controller *NotificationController) Subscribe() {
	param := controller.BuildQueryParam()
	param.Query["user_id"] = controller.User.ID
	IsReaded := controller.Input().Get("is_readed")
	if IsReaded != "" {

	}
	param.Relate = "all"
	var notificationLogs []models.NotificationLog
	err := models.GetAll(new(models.NotificationLog), &notificationLogs, param)
	if err != nil {

	}

	controller.Success(notificationLogs)
}

func (controller *NotificationController) Read() {

}
