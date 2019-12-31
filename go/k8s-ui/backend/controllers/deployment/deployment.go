package deployment

import "k8s-lx1036/k8s-ui/backend/controllers/base"

type DeploymentController struct {
	base.APIController
}

func (deployment *DeploymentController) URLMapping()  {
	deployment.Mapping("List", deployment.List)
	deployment.Mapping("Get", deployment.Get)
}

func (deployment *DeploymentController) Prepare() {

}

// @Param name query string false "name filter"
// @router / [get]
func (deployment *DeploymentController) List() {
	//params = deployment.BuildQueryParams()
	deployment.Input().Get("name")
}

func (deployment *DeploymentController) Get() {

}
