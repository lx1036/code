package permission

import (
	"k8s-lx1036/k8s-ui/backend/controllers/base"
	"net/http"
)

type AppUserController struct {
	base.APIController
}

func (c *AppUserController) URLMapping() {
	//c.Mapping("List", c.List)
	//c.Mapping("Create", c.Create)
	c.Mapping("Get", c.Get)
	//c.Mapping("GetPermissionByApp", c.GetPermissionByApp)
	//c.Mapping("Update", c.Update)
	c.Mapping("Delete", c.Delete)
}

// @Title Get
// @Description find Object by id
// @Param	id		path 	int	true		"the id you want to get"
// @Success 200 {object} models.AppUser success
// @router /:id [get]
func (c *AppUserController) Get() {
	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = base.Result{Data: "asdf"}
	c.ServeJSON()
}
