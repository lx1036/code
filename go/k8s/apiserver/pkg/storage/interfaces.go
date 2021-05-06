package storage

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
)

// Interface offers a common interface for object marshaling/unmarshaling operations and
// hides all the storage-related operations behind it.
type Interface interface {
	// Create adds a new object at a key unless it already exists. 'ttl' is time-to-live
	// in seconds (0 means forever). If no error is returned and out is not nil, out will be
	// set to the read value from database.
	Create(ctx context.Context, key string, obj, out runtime.Object, ttl uint64) error

	// Delete removes the specified key and returns the value that existed at that spot.
	// If key didn't exist, it will return NotFound storage error.
	Delete(ctx context.Context, key string, out runtime.Object, preconditions *Preconditions, validateDeletion ValidateObjectFunc) error

	// Watch begins watching the specified key. Events are decoded into API objects,
	// and any items selected by 'p' are sent down to returned watch.Interface.
	// resourceVersion may be used to specify what version to begin watching,
	// which should be the current resourceVersion, and no longer rv+1
	// (e.g. reconnecting without missing any updates).
	// If resource version is "0", this interface will get current object at given key
	// and send it in an "ADDED" event, before watch starts.
	Watch(ctx context.Context, key string, opts ListOptions) (watch.Interface, error)

	// WatchList begins watching the specified key's items. Items are decoded into API
	// objects and any item selected by 'p' are sent down to returned watch.Interface.
	// resourceVersion may be used to specify what version to begin watching,
	// which should be the current resourceVersion, and no longer rv+1
	// (e.g. reconnecting without missing any updates).
	// If resource version is "0", this interface will list current objects directory defined by key
	// and send them in "ADDED" events, before watch starts.
	WatchList(ctx context.Context, key string, opts ListOptions) (watch.Interface, error)

	// Get unmarshals json found at key into objPtr. On a not found error, will either
	// return a zero object of the requested type, or an error, depending on 'opts.ignoreNotFound'.
	// Treats empty responses and nil response nodes exactly like a not found error.
	// The returned contents may be delayed, but it is guaranteed that they will
	// match 'opts.ResourceVersion' according 'opts.ResourceVersionMatch'.
	Get(ctx context.Context, key string, opts GetOptions, objPtr runtime.Object) error

	// List unmarshalls jsons found at directory defined by key and opaque them
	// into *List api object (an object that satisfies runtime.IsList definition).
	// The returned contents may be delayed, but it is guaranteed that they will
	// match 'opts.ResourceVersion' according 'opts.ResourceVersionMatch'.
	List(ctx context.Context, key string, opts ListOptions, listObj runtime.Object) error
}

// ListOptions provides the options that may be provided for storage list operations.
type ListOptions struct {
	// ResourceVersion provides a resource version constraint to apply to the list operation
	// as a "not older than" constraint: the result contains data at least as new as the provided
	// ResourceVersion. The newest available data is preferred, but any data not older than this
	// ResourceVersion may be served.
	ResourceVersion string
	// ResourceVersionMatch provides the rule for how the resource version constraint applies. If set
	// to the default value "" the legacy resource version semantic apply.
	ResourceVersionMatch metav1.ResourceVersionMatch
	// Predicate provides the selection rules for the list operation.
	Predicate SelectionPredicate
}

// GetOptions provides the options that may be provided for storage get operations.
type GetOptions struct {
	// IgnoreNotFound determines what is returned if the requested object is not found. If
	// true, a zero object is returned. If false, an error is returned.
	IgnoreNotFound bool
	// ResourceVersion provides a resource version constraint to apply to the get operation
	// as a "not older than" constraint: the result contains data at least as new as the provided
	// ResourceVersion. The newest available data is preferred, but any data not older than this
	// ResourceVersion may be served.
	ResourceVersion string
}

// Preconditions must be fulfilled before an operation (update, delete, etc.) is carried out.
type Preconditions struct {
	// Specifies the target UID.
	// +optional
	UID *types.UID `json:"uid,omitempty"`
	// Specifies the target ResourceVersion
	// +optional
	ResourceVersion *string `json:"resourceVersion,omitempty"`
}

// ValidateObjectFunc is a function to act on a given object. An error may be returned
// if the hook cannot be completed. The function may NOT transform the provided
// object.
type ValidateObjectFunc func(ctx context.Context, obj runtime.Object) error

// Versioner abstracts setting and retrieving metadata fields from database response
// onto the object ot list. It is required to maintain storage invariants - updating an
// object twice with the same data except for the ResourceVersion and SelfLink must be
// a no-op. A resourceVersion of type uint64 is a 'raw' resourceVersion,
// intended to be sent directly to or from the backend. A resourceVersion of
// type string is a 'safe' resourceVersion, intended for consumption by users.
type Versioner interface {
	// UpdateObject sets storage metadata into an API object. Returns an error if the object
	// cannot be updated correctly. May return nil if the requested object does not need metadata
	// from database.
	UpdateObject(obj runtime.Object, resourceVersion uint64) error
	// UpdateList sets the resource version into an API list object. Returns an error if the object
	// cannot be updated correctly. May return nil if the requested object does not need metadata from
	// database. continueValue is optional and indicates that more results are available if the client
	// passes that value to the server in a subsequent call. remainingItemCount indicates the number
	// of remaining objects if the list is partial. The remainingItemCount field is omitted during
	// serialization if it is set to nil.
	UpdateList(obj runtime.Object, resourceVersion uint64, continueValue string, remainingItemCount *int64) error
	// PrepareObjectForStorage should set SelfLink and ResourceVersion to the empty value. Should
	// return an error if the specified object cannot be updated.
	PrepareObjectForStorage(obj runtime.Object) error
	// ObjectResourceVersion returns the resource version (for persistence) of the specified object.
	// Should return an error if the specified object does not have a persistable version.
	ObjectResourceVersion(obj runtime.Object) (uint64, error)

	// ParseResourceVersion takes a resource version argument and
	// converts it to the storage backend. For watch we should pass to helper.Watch().
	// Because resourceVersion is an opaque value, the default watch
	// behavior for non-zero watch is to watch the next value (if you pass
	// "1", you will see updates from "2" onwards).
	ParseResourceVersion(resourceVersion string) (uint64, error)
}
