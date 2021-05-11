package scheme

import (
	"k8s-lx1036/k8s/apiserver/aggregator-server/pkg/apis/apiregistration"
	"k8s-lx1036/k8s/apiserver/aggregator-server/pkg/apis/apiregistration/install"
	"k8s-lx1036/k8s/apiserver/aggregator-server/pkg/apis/apiregistration/v1"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
)

var (
	// Scheme defines methods for serializing and deserializing API objects.
	Scheme = runtime.NewScheme()
	// Codecs provides methods for retrieving codecs and serializers for specific
	// versions and content types.
	Codecs = serializer.NewCodecFactory(Scheme)
)

func init() {
	AddToScheme(Scheme)
	install.Install(Scheme)
}

// AddToScheme adds the types of this group into the given scheme.
func AddToScheme(scheme *runtime.Scheme) {
	utilruntime.Must(v1.AddToScheme(scheme))
	utilruntime.Must(apiregistration.AddToScheme(scheme))
}
