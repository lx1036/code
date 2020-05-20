package deployment

import (
	apps "k8s.io/api/apps/v1"
	api "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	core "k8s.io/client-go/testing"
	"reflect"
	"testing"
)

func TestDeployApp(test *testing.T) {
	replicas := int32(0)
	namespace := "foo-namespace"
	spec := &AppDeploymentSpec{
		Namespace:       namespace,
		Name:            "foo-name",
		RunAsPrivileged: true,
		Replicas:        replicas,
	}

	expected := &apps.Deployment{
		ObjectMeta: v1.ObjectMeta{
			Name:        "foo-name",
			Labels:      map[string]string{},
			Annotations: map[string]string{},
		},
		Spec: apps.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{},
			},
			Template: api.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:        "foo-name",
					Labels:      map[string]string{},
					Annotations: map[string]string{},
				},
				Spec: api.PodSpec{
					Containers: []api.Container{
						{
							Name: "foo-bar",
							SecurityContext: &api.SecurityContext{
								Privileged: &spec.RunAsPrivileged,
							},
							Resources: api.ResourceRequirements{
								Requests: api.ResourceList{},
							},
						},
					},
				},
			},
		},
	}

	testClient := fake.NewSimpleClientset()
	_ = DeployApp(spec, testClient)

	createAction := testClient.Actions()[0].(core.CreateActionImpl)
	if len(testClient.Actions()) != 1 {
		test.Errorf("Expected one create action but got %#v", len(testClient.Actions()))
	}
	if createAction.GetNamespace() != namespace {
		test.Errorf("Expected namespace to be %#v but go %#v", namespace, createAction.GetNamespace())
	}
	deployment := createAction.GetObject().(*apps.Deployment)
	if !reflect.DeepEqual(deployment, expected) {
		test.Errorf("Expected replication controller \n%#v\n to be created but got \n%#v\n", expected, deployment)
	}
}

func TestDeployAppContainerCommands(test *testing.T) {
}
