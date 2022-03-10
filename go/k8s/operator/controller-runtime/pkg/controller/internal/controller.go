package internal

import (
	"k8s-lx1036/k8s/concepts/tools/kubebuilder/controller-runtime/pkg/controller/reconcile"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/source"
	"sync"
)

type Controller struct {
	Name string
	
	MaxConcurrentReconciles int
	
	Do reconcile.Reconciler
	
	MakeQueue func() workqueue.RateLimitingInterface
	
	Queue workqueue.RateLimitingInterface
	
	mu sync.Mutex
	
	//Log logr.Logger
	
	SetFields func(i interface{}) error
}

func (c *Controller) Reconcile(r reconcile.Request) (reconcile.Result, error) {
	return c.Do.Reconcile(r)
}

func (c *Controller) Watch(src source.Source, evthdler handler.EventHandler, prct ...predicate.Predicate) error {

}

func (c *Controller) Start(stop <-chan struct{}) error {

}
