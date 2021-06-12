package scheme

import (
	v1 "k8s-lx1036/k8s/bigdata/spark-on-k8s/spark-operator/pkg/apis/sparkoperator.k9s.io/v1"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes/scheme"
)

// INFO: 本来打算在main.go里import下，但是eventBroadcaster.NewRecorder(scheme.Scheme)使用的是这个scheme.Scheme，
//  所以v1.AddToScheme必须也要注册到这个scheme.Scheme里。
//  但是决定不这么做，直接在controller里一行代码直接注册

var (
	// INFO: 这里注册到 scheme.Scheme，这个Scheme包含了k8s.io/api内置的所有对象
	Scheme = scheme.Scheme
)

func init() {
	AddToScheme(Scheme)
}

// AddToScheme builds the kubescheduler scheme using all known versions of the kubescheduler api.
func AddToScheme(scheme *runtime.Scheme) {
	utilruntime.Must(v1.AddToScheme(scheme))
}
