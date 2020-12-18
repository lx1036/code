package controller

import (
	"k8s-lx1036/k8s/concepts/components/controller-manager/calico-node-controller/pkg/calico"
	"k8s-lx1036/k8s/concepts/components/controller-manager/calico-node-controller/pkg/kube"
)

type NodeController struct {
}

func NewNodeController() *NodeController {
	kubeClientset := kube.GetKubernetesClientset()
	calicoClient := calico.GetCalicoClientOrDie()

}

func (controller *NodeController) Run(workers int, stopCh <-chan struct{}) error {

}
