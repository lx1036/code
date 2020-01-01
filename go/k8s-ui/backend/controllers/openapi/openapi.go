package openapi

import "k8s-lx1036/k8s-ui/backend/controllers/base"

const (
	UpgradeDeploymentAction = "UPGRADE_DEPLOYMENT"
	GetPodListAction        = "GET_POD_LIST"
)

type OpenAPIController struct {
	base.APIKeyController
}

func (controller *OpenAPIController) Prepare() {
	controller.APIKeyController.Prepare()
}

func (controller *OpenAPIController) CheckoutRoutePermission(action string) bool {
	return true
}

func (controller *OpenAPIController) CheckDeploymentPermission(deployment string) bool {
	return true
}

func (controller *OpenAPIController) CheckNamespacePermission(namespace string) bool {
	return true
}
