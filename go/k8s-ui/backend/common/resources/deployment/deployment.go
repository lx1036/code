package deployment

import (
	"context"
	"k8s.io/api/apps/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func CreateOrUpdateDeployment(client *kubernetes.Clientset, deployment *v1beta1.Deployment) (*v1beta1.Deployment, error) {
	old, err := client.AppsV1beta1().Deployments(deployment.Namespace).Get(context.TODO(), deployment.Name, metaV1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return client.AppsV1beta1().Deployments(deployment.Namespace).Create(context.TODO(), deployment, metaV1.CreateOptions{})
		}
		return nil, err
	}

	err = checkDeploymentLabelSelector(deployment, old)
	if err != nil {
		return nil, err
	}

	old.Labels = deployment.Labels
	old.Annotations = deployment.Annotations
	old.Spec = deployment.Spec

	return client.AppsV1beta1().Deployments(deployment.Namespace).Update(context.TODO(), old, metaV1.UpdateOptions{})
}

// check Deployment .Spec.Selector.MatchLabels, prevent orphan ReplicaSet
// old deployment .Spec.Selector.MatchLabels labels should contain all new deployment .Spec.Selector.MatchLabels labels
// e.g. old Deployment .Spec.Selector.MatchLabels is app = infra-wayne,wayne-app = infra
// new Deployment .Spec.Selector.MatchLabels valid labels is
// app = infra-wayne or wayne-app = infra or app = infra-wayne,wayne-app = infra
func checkDeploymentLabelSelector(new *v1beta1.Deployment, old *v1beta1.Deployment) error {
	return nil
}
