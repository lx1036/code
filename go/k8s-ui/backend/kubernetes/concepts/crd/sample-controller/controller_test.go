package main

import (
	samplecontroller "k8s-lx1036/k8s-ui/backend/kubernetes/crd/sample-controller/pkg/apis/samplecontroller/v1alpha1"
	"k8s-lx1036/k8s-ui/backend/kubernetes/crd/sample-controller/pkg/generated/clientset/versioned/fake"
	"k8s.io/apimachinery/pkg/runtime"
	k8sfake "k8s.io/client-go/kubernetes/fake"
	core "k8s.io/client-go/testing"
	"k8s.io/kubernetes/pkg/apis/apps"
	"testing"
)

type Fixture struct {
	t *testing.T

	client     *fake.Clientset
	kubeclient *k8sfake.Clientset
	// Objects to put in the store.
	fooLister        []*samplecontroller.Foo
	deploymentLister []*apps.Deployment
	// Actions expected to happen on the client.
	kubeactions []core.Action
	actions     []core.Action
	// Objects from here preloaded into NewSimpleFake.
	kubeobjects []runtime.Object
	objects     []runtime.Object
}

func newFixture(test *testing.T) *Fixture {

}

func TestCreatesDeployment(test *testing.T) {
	fixture := newFixture(test)
}

func TestDoNothing(test *testing.T) {

}

func TestUpdateDeployment(test *testing.T) {

}

func TestNotControlledByUs(test *testing.T) {

}
