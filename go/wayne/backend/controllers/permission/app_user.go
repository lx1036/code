package permission

import (
	"k8s-lx1036/wayne/backend/controllers/base"
)

type AppUserController struct {
	base.APIController
}

func (c *AppUserController) URLMapping()  {
	c.Mapping("List", c.List)
	c.Mapping("Create", c.Create)
	c.Mapping("Get", c.Get)
	c.Mapping("GetPermissionByApp", c.GetPermissionByApp)
	c.Mapping("Update", c.Update)
	c.Mapping("Delete", c.Delete)
}



