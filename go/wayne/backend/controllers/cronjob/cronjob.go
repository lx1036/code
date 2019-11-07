package cronjob

import (
	"github.com/astaxie/beego/logs"
	"k8s-lx1036/wayne/backend/controllers/base"
	"k8s-lx1036/wayne/backend/models"
)

type CronjobController struct {
	base.APIController
}

func (c *CronjobController) URLMapping() {
	c.Mapping("GetNames", c.GetNames)
	c.Mapping("List", c.List)
	c.Mapping("Create", c.Create)
	c.Mapping("Get", c.Get)
	c.Mapping("Update", c.Update)
	c.Mapping("Delete", c.Delete)
}

func (c *CronjobController) Prepare() {
	// Check administration
	c.APIController.Prepare()

	// Check permission
	perAction := ""
	_, method := c.GetControllerAndAction()
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
		c.CheckPermission(models.PermissionTypeCronjob, perAction)
	}
}

func (c *CronjobController) GetNames() {

}

func (c *CronjobController) Get() {
	id := c.GetIDFromURL()
	cronjob, err := models.CronjobModel.GetById(int64(id))
	if err != nil {
		logs.Error("get by id (%d) error.%v", id, err)
		c.HandleError(err)
		return
	}

	c.Success(cronjob)
	return
}
