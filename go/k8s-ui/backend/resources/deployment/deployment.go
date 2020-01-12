package deployment

import (
	"k8s.io/api/apps/v1beta1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func CreateOrUpdateDeployment(cli *kubernetes.Clientset, deployment *v1beta1.Deployment) (*v1beta1.Deployment, error) {
	old, err := cli.AppsV1beta1().Deployments(deployment.Namespace).Get(deployment.Name, metaV1.GetOptions{})
	if err != nil {

	}

	old.Labels = deployment.Labels
	old.Annotations = deployment.Annotations
	old.Spec = deployment.Spec

	return cli.AppsV1beta1().Deployments(deployment.Namespace).Update(old)
}
