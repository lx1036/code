package permission

import (
	"github.com/mitchellh/mapstructure"
	"k8s-lx1036/k8s-ui/backend/controllers/base"
	"k8s-lx1036/k8s-ui/backend/models"
)

type NamespaceUserController struct {
	base.APIController
}

func (controller *NamespaceUserController) URLMapping() {
	controller.Mapping("GetPermissionByNS", controller.GetPermissionByNS)
}

// @Title GetPerNS
// @Description get PerNS by nsId
// @Param	id		path 	int	true		"the ns id"
// @Success 200 {object} models.TypeApp success
// @router /permissions/:id [get]
func (controller *NamespaceUserController) GetPermissionByNS() {
	id := controller.GetIdFromURL()
	permissions, err := models.NamespaceUserModel.GetAllPermissions(id, controller.User.Id)
	if err != nil {

	}

	mapPer := make(map[string]map[string]bool)
	for _, permission := range permissions {
		paction, ptype, err := models.PermissionModel.SplitName(permission.Name)
		if err != nil {

		}
		mapPer[ptype][paction] = true
	}

	var ret models.TypePermission
	if err := mapstructure.Decode(mapPer, &ret); err != nil {

	}

	controller.Success(ret)
}
