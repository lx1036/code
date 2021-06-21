package targetloadpacking

import (
	"k8s.io/apimachinery/pkg/runtime"
	framework "k8s.io/kubernetes/pkg/scheduler/framework/v1alpha1"

)

func New(obj runtime.Object, handle framework.FrameworkHandle) (framework.Plugin, error) {

}

