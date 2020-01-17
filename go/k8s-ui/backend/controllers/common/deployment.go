package common

import (
	"k8s-lx1036/k8s-ui/backend/models"
	"k8s.io/api/apps/v1beta1"
)

func DeploymentPreDeploy(kubeDeployment *v1beta1.Deployment, deploy *models.Deployment,
	cluster *models.Cluster, namespace *models.Namespace) {

}
