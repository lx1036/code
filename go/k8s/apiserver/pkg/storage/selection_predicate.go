package storage

import (
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
)

// AttrFunc returns label and field sets and the uninitialized flag for List or Watch to match.
// In any failure to parse given object, it returns error.
type AttrFunc func(obj runtime.Object) (labels.Set, fields.Set, error)

// SelectionPredicate is used to represent the way to select objects from api storage.
type SelectionPredicate struct {
	Label               labels.Selector
	Field               fields.Selector
	GetAttrs            AttrFunc
	IndexLabels         []string
	IndexFields         []string
	Limit               int64
	Continue            string
	AllowWatchBookmarks bool
}
