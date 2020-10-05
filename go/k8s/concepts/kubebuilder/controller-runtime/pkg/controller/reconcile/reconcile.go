package reconcile

import (
	"k8s.io/apimachinery/pkg/types"
	"time"
)

type Request struct {
	// NamespacedName is the name and namespace of the object to reconcile.
	types.NamespacedName
}

type Result struct {
	// Requeue tells the Controller to requeue the reconcile key.  Defaults to false.
	Requeue bool

	// RequeueAfter if greater than 0, tells the Controller to requeue the reconcile key after the Duration.
	// Implies that Requeue is true, there is no need to set Requeue to true at the same time as RequeueAfter.
	RequeueAfter time.Duration
}

type Reconciler interface {
	// Reconciler performs a full reconciliation for the object referred to by the Request.
	// The Controller will requeue the Request to be processed again if an error is non-nil or
	// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
	Reconcile(Request) (Result, error)
}

// Func is a function that implements the reconcile interface.
type Func func(Request) (Result, error)

var _ Reconciler = Func(nil)

// Reconcile implements Reconciler.
func (r Func) Reconcile(o Request) (Result, error) { return r(o) }
