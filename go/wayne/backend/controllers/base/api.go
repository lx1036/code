package base

import (
	"k8s-lx1036/wayne/backend/models"
	"net/http"
)

type APIController struct {
	LoggedInController

	NamespaceId int64
	AppId       int64
}

// Abort stops controller handler and show the error data， e.g. Prepare
func (c *APIController) AbortForbidden(msg string) {
	c.publishRequestMessage(http.StatusForbidden, msg)

	c.ResultHandlerController.AbortForbidden(msg)
}

/*
 * 检查资源权限
 */
func (c *APIController) CheckPermission(perType string, perAction string) {
	// 如果用户是admin，跳过permission检查
	if c.User.Admin {
		return
	}

	perName := models.PermissionModel.MergeName(perType, perAction)
	if c.NamespaceId != 0 {

	}

	c.AbortForbidden("Permission error")

}
