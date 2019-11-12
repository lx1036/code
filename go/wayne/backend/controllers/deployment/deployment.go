package deployment

import "k8s-lx1036/wayne/backend/controllers/base"

type DeploymentController struct {
	base.APIController
}

func (deployment *DeploymentController) URLMapping()  {
	deployment.Mapping("List", deployment.List)
}

func (deployment *DeploymentController) List() {

}
