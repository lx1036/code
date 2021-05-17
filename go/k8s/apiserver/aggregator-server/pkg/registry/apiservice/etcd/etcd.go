package etcd

import (
	"k8s-lx1036/k8s/apiserver/aggregator-server/pkg/apis/apiregistration"
	"k8s-lx1036/k8s/apiserver/aggregator-server/pkg/registry/apiservice"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apiserver/pkg/registry/generic"
	genericregistry "k8s.io/apiserver/pkg/registry/generic/registry"
	"k8s.io/apiserver/pkg/registry/rest"
)

// REST implements a RESTStorage for API services against etcd
type REST struct {
	*genericregistry.Store
}

// NewREST returns a RESTStorage object that will work against API services.
func NewREST(scheme *runtime.Scheme, optsGetter generic.RESTOptionsGetter) *REST {
	strategy := apiservice.NewStrategy(scheme)
	store := &genericregistry.Store{
		NewFunc:                  func() runtime.Object { return &apiregistration.APIService{} },
		NewListFunc:              func() runtime.Object { return &apiregistration.APIServiceList{} },
		PredicateFunc:            apiservice.MatchAPIService,
		DefaultQualifiedResource: apiregistration.Resource("apiservices"),

		CreateStrategy: strategy,
		UpdateStrategy: strategy,
		DeleteStrategy: strategy,

		// TODO: define table converter that exposes more than name/creation timestamp
		TableConvertor: rest.NewDefaultTableConvertor(apiregistration.Resource("apiservices")),
	}

	options := &generic.StoreOptions{RESTOptions: optsGetter, AttrFunc: apiservice.GetAttrs}
	if err := store.CompleteWithOptions(options); err != nil {
		panic(err) // TODO: Propagate error up
	}

	return &REST{store}
}

// StatusREST implements the REST endpoint for changing the status of an APIService.
type StatusREST struct {
	store *genericregistry.Store
}

// New creates a new APIService object.
func (s StatusREST) New() runtime.Object {
	return &apiregistration.APIService{}
}

// NewStatusREST makes a RESTStorage for status that has more limited options.
// It is based on the original REST so that we can share the same underlying store
func NewStatusREST(scheme *runtime.Scheme, rest *REST) *StatusREST {
	statusStore := *rest.Store
	statusStore.CreateStrategy = nil
	statusStore.DeleteStrategy = nil
	statusStore.UpdateStrategy = apiservice.NewStatusStrategy(scheme)
	return &StatusREST{store: &statusStore}
}
