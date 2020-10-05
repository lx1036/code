package client

import (
	"context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
)

// InNamespace restricts the list/delete operation to the given namespace.
type InNamespace string

// ApplyToList applies this configuration to the given list options.
func (n InNamespace) ApplyToList(opts *ListOptions) {
	opts.Namespace = string(n)
}

// ApplyToDeleteAllOf applies this configuration to the given an List options.
func (n InNamespace) ApplyToDeleteAllOf(opts *DeleteAllOfOptions) {
	n.ApplyToList(&opts.ListOptions)
}

type ListOption interface {
	ApplyToList(*ListOptions)
}

type ListOptions struct {
	LabelSelector labels.Selector

	FieldSelector fields.Selector

	Namespace string

	Limit int64

	Continue string

	Raw *metav1.ListOptions
}

// ApplyOptions applies the given list options on these options,
// and then returns itself (for convenient chaining).
func (o *ListOptions) ApplyOptions(opts []ListOption) *ListOptions {
	for _, opt := range opts {
		opt.ApplyToList(o)
	}
	return o
}

// AsListOptions returns these options as a flattened metav1.ListOptions.
// This may mutate the Raw field.
func (o *ListOptions) AsListOptions() *metav1.ListOptions {
	if o == nil {
		return &metav1.ListOptions{}
	}
	if o.Raw == nil {
		o.Raw = &metav1.ListOptions{}
	}
	if o.LabelSelector != nil {
		o.Raw.LabelSelector = o.LabelSelector.String()
	}
	if o.FieldSelector != nil {
		o.Raw.FieldSelector = o.FieldSelector.String()
	}
	if !o.Raw.Watch {
		o.Raw.Limit = o.Limit
		o.Raw.Continue = o.Continue
	}
	return o.Raw
}

type CreateOption interface {
	ApplyToCreate(*CreateOptions)
}

type CreateOptions struct {
	DryRun []string

	FieldManager string

	Raw *metav1.CreateOptions
}

type DeleteOption interface {
	ApplyToDelete(*DeleteOptions)
}

type DeleteOptions struct {
	GracePeriodSeconds *int64

	Preconditions *metav1.Preconditions

	PropagationPolicy *metav1.DeletionPropagation

	Raw *metav1.DeleteOptions

	DryRun []string
}

type UpdateOption interface {
	ApplyToUpdate(*UpdateOptions)
}

type UpdateOptions struct {
	DryRun []string

	FieldManager string

	Raw *metav1.UpdateOptions
}

type PatchOption interface {
	ApplyToPatch(*PatchOptions)
}

type PatchOptions struct {
	DryRun []string

	Force *bool

	FieldManager string

	Raw *metav1.PatchOptions
}

type DeleteAllOfOption interface {
	ApplyToDeleteAllOf(*DeleteAllOfOptions)
}

type DeleteAllOfOptions struct {
	ListOptions
	DeleteOptions
}

type StatusClient interface {
	Status() StatusWriter
}

type StatusWriter interface {
	Update(ctx context.Context, obj runtime.Object, opts ...UpdateOption) error

	Patch(ctx context.Context, obj runtime.Object, patch Patch, opts ...PatchOption) error
}
