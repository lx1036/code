package install

import (
	"k8s-lx1036/k8s/apiserver/aggregator-server/pkg/apis/apiregistration"
	"k8s-lx1036/k8s/apiserver/aggregator-server/pkg/apis/apiregistration/v1"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
)

// Install registers the API group and adds types to a scheme
func Install(scheme *runtime.Scheme) {
	utilruntime.Must(apiregistration.AddToScheme(scheme))
	utilruntime.Must(v1.AddToScheme(scheme))
	utilruntime.Must(scheme.SetVersionPriority(v1.SchemeGroupVersion))
}
