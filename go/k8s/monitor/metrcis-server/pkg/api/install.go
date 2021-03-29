package api

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apiserver/pkg/registry/rest"
	genericapiserver "k8s.io/apiserver/pkg/server"
	corev1 "k8s.io/client-go/informers/core/v1"
	"k8s.io/metrics/pkg/apis/metrics"
	"k8s.io/metrics/pkg/apis/metrics/install"
	"k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

var (
	// Scheme contains the types needed by the resource metrics API.
	Scheme = runtime.NewScheme()
	// Codecs is a codec factory for serving the resource metrics API.
	Codecs = serializer.NewCodecFactory(Scheme)
)

func init() {
	install.Install(Scheme)
	metav1.AddToGroupVersion(Scheme, schema.GroupVersion{Version: "v1"})
}

// Build constructs APIGroupInfo the metrics.k8s.io API group using the given getters.
func Build(m MetricsGetter, informers corev1.Interface) genericapiserver.APIGroupInfo {
	apiGroupInfo := genericapiserver.NewDefaultAPIGroupInfo(metrics.GroupName, Scheme, metav1.ParameterCodec, Codecs)

	node := newNodeMetrics(metrics.Resource("nodemetrics"), m, informers.Nodes().Lister())
	pod := newPodMetrics(metrics.Resource("podmetrics"), m, informers.Pods().Lister())
	metricsServerResources := map[string]rest.Storage{
		"nodes": node,
		"pods":  pod,
	}
	apiGroupInfo.VersionedResourcesStorageMap[v1beta1.SchemeGroupVersion.Version] = metricsServerResources

	return apiGroupInfo
}

// InstallStorage builds the metrics for the metrics.k8s.io API, and then installs it into the given API metrics-server.
func Install(metrics MetricsGetter, informers corev1.Interface, server *genericapiserver.GenericAPIServer) error {
	info := Build(metrics, informers)
	return server.InstallAPIGroup(&info)
}
