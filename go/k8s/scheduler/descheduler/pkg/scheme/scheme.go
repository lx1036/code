package scheme

import (
	"k8s-lx1036/k8s/scheduler/descheduler/pkg/api"
	apiv1alpha1 "k8s-lx1036/k8s/scheduler/descheduler/pkg/api/v1alpha1"
	"k8s-lx1036/k8s/scheduler/descheduler/pkg/apis/componentconfig"
	componentconfigv1alpha1 "k8s-lx1036/k8s/scheduler/descheduler/pkg/apis/componentconfig/v1alpha1"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
)

// 注册CRD到k8s scheme

var (
	Scheme = runtime.NewScheme()
	Codecs = serializer.NewCodecFactory(Scheme)
)

func init() {
	utilruntime.Must(api.AddToScheme(Scheme))
	utilruntime.Must(apiv1alpha1.AddToScheme(Scheme))

	utilruntime.Must(componentconfig.AddToScheme(Scheme))
	utilruntime.Must(componentconfigv1alpha1.AddToScheme(Scheme))
}
