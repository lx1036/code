package client

import (
	"context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
)

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
