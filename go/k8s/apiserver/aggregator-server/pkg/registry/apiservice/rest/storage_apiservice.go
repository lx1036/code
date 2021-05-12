package rest

import (
	"k8s-lx1036/k8s/apiserver/aggregator-server/pkg/apis/apiregistration"
	"k8s-lx1036/k8s/apiserver/aggregator-server/pkg/apis/apiregistration/v1"
	aggregatorscheme "k8s-lx1036/k8s/apiserver/aggregator-server/pkg/apiserver/scheme"
	apiservicestorage "k8s-lx1036/k8s/apiserver/aggregator-server/pkg/registry/apiservice/etcd"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apiserver/pkg/registry/generic"
	"k8s.io/apiserver/pkg/registry/rest"
	genericapiserver "k8s.io/apiserver/pkg/server"
	serverstorage "k8s.io/apiserver/pkg/server/storage"
)

// NewRESTStorage returns an APIGroupInfo object that will work against apiservice.
func NewRESTStorage(apiResourceConfigSource serverstorage.APIResourceConfigSource,
	restOptionsGetter generic.RESTOptionsGetter) genericapiserver.APIGroupInfo {
	apiGroupInfo := genericapiserver.NewDefaultAPIGroupInfo(apiregistration.GroupName, aggregatorscheme.Scheme,
		metav1.ParameterCodec, aggregatorscheme.Codecs)

	if apiResourceConfigSource.VersionEnabled(v1.SchemeGroupVersion) {
		storage := map[string]rest.Storage{}
		apiServiceREST := apiservicestorage.NewREST(aggregatorscheme.Scheme, restOptionsGetter)
		storage["apiservices"] = apiServiceREST
		storage["apiservices/status"] = apiservicestorage.NewStatusREST(aggregatorscheme.Scheme, apiServiceREST)
		apiGroupInfo.VersionedResourcesStorageMap["v1"] = storage
	}

	return apiGroupInfo
}
