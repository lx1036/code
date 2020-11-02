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

func (o *CreateOptions) ApplyOptions(opts []CreateOption) *CreateOptions {
	for _, opt := range opts {
		opt.ApplyToCreate(o)
	}
	return o
}

// AsCreateOptions returns these options as a metav1.CreateOptions.
// This may mutate the Raw field.
func (o *CreateOptions) AsCreateOptions() *metav1.CreateOptions {
	if o == nil {
		return &metav1.CreateOptions{}
	}
	if o.Raw == nil {
		o.Raw = &metav1.CreateOptions{}
	}

	o.Raw.DryRun = o.DryRun
	o.Raw.FieldManager = o.FieldManager
	return o.Raw
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

// AsDeleteOptions returns these options as a metav1.DeleteOptions.
// This may mutate the Raw field.
func (o *DeleteOptions) AsDeleteOptions() *metav1.DeleteOptions {
	if o == nil {
		return &metav1.DeleteOptions{}
	}
	if o.Raw == nil {
		o.Raw = &metav1.DeleteOptions{}
	}

	o.Raw.GracePeriodSeconds = o.GracePeriodSeconds
	o.Raw.Preconditions = o.Preconditions
	o.Raw.PropagationPolicy = o.PropagationPolicy
	o.Raw.DryRun = o.DryRun
	return o.Raw
}

// ApplyOptions applies the given delete options on these options,
// and then returns itself (for convenient chaining).
func (o *DeleteOptions) ApplyOptions(opts []DeleteOption) *DeleteOptions {
	for _, opt := range opts {
		opt.ApplyToDelete(o)
	}
	return o
}

type UpdateOption interface {
	ApplyToUpdate(*UpdateOptions)
}

type UpdateOptions struct {
	DryRun []string

	FieldManager string

	Raw *metav1.UpdateOptions
}

// ApplyOptions applies the given update options on these options,
// and then returns itself (for convenient chaining).
func (o *UpdateOptions) ApplyOptions(opts []UpdateOption) *UpdateOptions {
	for _, opt := range opts {
		opt.ApplyToUpdate(o)
	}
	return o
}

// AsUpdateOptions returns these options as a metav1.UpdateOptions.
// This may mutate the Raw field.
func (o *UpdateOptions) AsUpdateOptions() *metav1.UpdateOptions {
	if o == nil {
		return &metav1.UpdateOptions{}
	}
	if o.Raw == nil {
		o.Raw = &metav1.UpdateOptions{}
	}

	o.Raw.DryRun = o.DryRun
	o.Raw.FieldManager = o.FieldManager
	return o.Raw
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

func (o *PatchOptions) ApplyOptions(opts []PatchOption) *PatchOptions {
	for _, opt := range opts {
		opt.ApplyToPatch(o)
	}
	return o
}
func (o *PatchOptions) AsPatchOptions() *metav1.PatchOptions {
	if o == nil {
		return &metav1.PatchOptions{}
	}
	if o.Raw == nil {
		o.Raw = &metav1.PatchOptions{}
	}

	o.Raw.DryRun = o.DryRun
	o.Raw.Force = o.Force
	o.Raw.FieldManager = o.FieldManager
	return o.Raw
}

type DeleteAllOfOption interface {
	ApplyToDeleteAllOf(*DeleteAllOfOptions)
}

type DeleteAllOfOptions struct {
	ListOptions
	DeleteOptions
}

func (o *DeleteAllOfOptions) ApplyOptions(opts []DeleteAllOfOption) *DeleteAllOfOptions {
	for _, opt := range opts {
		opt.ApplyToDeleteAllOf(o)
	}
	return o
}

type StatusClient interface {
	Status() StatusWriter
}

type StatusWriter interface {
	Update(ctx context.Context, obj runtime.Object, opts ...UpdateOption) error

	Patch(ctx context.Context, obj runtime.Object, patch Patch, opts ...PatchOption) error
}

type MatchingLabels map[string]string

func (m MatchingLabels) ApplyToDeleteAllOf(options *DeleteAllOfOptions) {
	m.ApplyToList(&options.ListOptions)
}
func (m MatchingLabels) ApplyToList(opts *ListOptions) {
	// TODO(directxman12): can we avoid reserializing this over and over?
	sel := labels.SelectorFromValidatedSet(map[string]string(m))
	opts.LabelSelector = sel
}
