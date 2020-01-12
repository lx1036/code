package base

import (
	"net/http"
)

type APIController struct {
	LoggedInController

	NamespaceId int64
	AppId       int64
}

// Abort stops controller handler and show the error data， e.g. Prepare
func (controller *APIController) AbortForbidden(msg string) {
	//c.publishRequestMessage(http.StatusForbidden, msg)

	controller.ResultHandlerController.AbortForbidden(msg)
}

/*
 * 检查资源权限
 */
func (controller *APIController) CheckPermission(perType string, perAction string) {
	// 如果用户是admin，跳过permission检查
	if controller.User.Admin {
		return
	}

	//perName := models.PermissionModel.MergeName(perType, perAction)
	if controller.NamespaceId != 0 {

	}

	controller.AbortForbidden("Permission error")

}

func (controller *APIController) Success(data interface{}) {
	controller.publishRequestMessage(http.StatusOK, data)
	controller.ResultHandlerController.Success(data)
}
