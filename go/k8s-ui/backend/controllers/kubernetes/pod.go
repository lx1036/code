package kubernetes

import "k8s-lx1036/k8s-ui/backend/controllers/base"

type KubePodController struct {
	base.APIController
}

func (controller *KubePodController) URLMapping() {
	controller.Mapping("PodStatistics", controller.PodStatistics)

}

func (controller *KubePodController) PodStatistics() {

}
