package apiservice

import (
	"context"
	"fmt"

	"k8s-lx1036/k8s/apiserver/aggregator-server/pkg/apis/apiregistration"

	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/apiserver/pkg/registry/generic"
	"k8s.io/apiserver/pkg/registry/rest"
	"k8s.io/apiserver/pkg/storage"
	"k8s.io/apiserver/pkg/storage/names"
)

type apiServerStrategy struct {
	runtime.ObjectTyper
	names.NameGenerator
}

func (apiServerStrategy) NamespaceScoped() bool {
	panic("implement me")
}

func (apiServerStrategy) PrepareForCreate(ctx context.Context, obj runtime.Object) {
	panic("implement me")
}

func (apiServerStrategy) Validate(ctx context.Context, obj runtime.Object) field.ErrorList {
	panic("implement me")
}

func (apiServerStrategy) Canonicalize(obj runtime.Object) {
	panic("implement me")
}

func (apiServerStrategy) AllowCreateOnUpdate() bool {
	panic("implement me")
}

func (apiServerStrategy) PrepareForUpdate(ctx context.Context, obj, old runtime.Object) {
	panic("implement me")
}

func (apiServerStrategy) ValidateUpdate(ctx context.Context, obj, old runtime.Object) field.ErrorList {
	panic("implement me")
}

func (apiServerStrategy) AllowUnconditionalUpdate() bool {
	panic("implement me")
}

// NewStrategy creates a new apiServerStrategy.
func NewStrategy(typer runtime.ObjectTyper) rest.RESTCreateUpdateStrategy {
	return apiServerStrategy{typer, names.SimpleNameGenerator}
}

type apiServerStatusStrategy struct {
	runtime.ObjectTyper
	names.NameGenerator
}

func (apiServerStatusStrategy) NamespaceScoped() bool {
	panic("implement me")
}

func (apiServerStatusStrategy) AllowCreateOnUpdate() bool {
	panic("implement me")
}

func (apiServerStatusStrategy) PrepareForUpdate(ctx context.Context, obj, old runtime.Object) {
	panic("implement me")
}

func (apiServerStatusStrategy) ValidateUpdate(ctx context.Context, obj, old runtime.Object) field.ErrorList {
	panic("implement me")
}

func (apiServerStatusStrategy) Canonicalize(obj runtime.Object) {
	panic("implement me")
}

func (apiServerStatusStrategy) AllowUnconditionalUpdate() bool {
	panic("implement me")
}

// NewStatusStrategy creates a new apiServerStatusStrategy.
func NewStatusStrategy(typer runtime.ObjectTyper) rest.RESTUpdateStrategy {
	return apiServerStatusStrategy{typer, names.SimpleNameGenerator}
}

// GetAttrs returns the labels and fields of an API server for filtering purposes.
func GetAttrs(obj runtime.Object) (labels.Set, fields.Set, error) {
	apiserver, ok := obj.(*apiregistration.APIService)
	if !ok {
		return nil, nil, fmt.Errorf("given object is not a APIService")
	}

	return labels.Set(apiserver.ObjectMeta.Labels), ToSelectableFields(apiserver), nil
}

// ToSelectableFields returns a field set that represents the object.
func ToSelectableFields(obj *apiregistration.APIService) fields.Set {
	return generic.ObjectMetaFieldsSet(&obj.ObjectMeta, true)
}

// MatchAPIService is the filter used by the generic etcd backend to watch events
// from etcd to clients of the apiserver only interested in specific labels/fields.
func MatchAPIService(label labels.Selector, field fields.Selector) storage.SelectionPredicate {
	return storage.SelectionPredicate{
		Label:    label,
		Field:    field,
		GetAttrs: GetAttrs,
	}
}
