package openapi

import "k8s-lx1036/wayne/backend/controllers/base"


const (
	UpgradeDeploymentAction      = "UPGRADE_DEPLOYMENT"
)

type OpenAPIController struct {
	base.APIKeyController
}

func (c *OpenAPIController) Prepare() {
	c.APIKeyController.Prepare()
}

func (c *OpenAPIController) CheckoutRoutePermission(action string) bool  {
	return true
}


func (c *OpenAPIController) CheckDeploymentPermission(deployment string) bool  {
	return true
}

func (c *OpenAPIController) CheckNamespacePermission(namespace string) bool  {
	return true
}
