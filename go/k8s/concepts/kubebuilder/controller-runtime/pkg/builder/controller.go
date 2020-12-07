package builder

import (
	"fmt"
	"k8s-lx1036/k8s/concepts/kubebuilder/controller-runtime/pkg/controller"
	"k8s-lx1036/k8s/concepts/kubebuilder/controller-runtime/pkg/manager"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"strings"
)

type Builder struct {
	mgr  manager.Manager
	name string

	forInput    ForInput
	ownsInput   []OwnsInput
	config      *rest.Config
	ctrl        controller.Controller
	ctrlOptions controller.Options
}

type ForOption interface {
	ApplyToFor(*ForInput)
}
type ForInput struct {
	object     runtime.Object
	predicates []predicate.Predicate
}
type OwnsOption interface {
	ApplyToOwns(*OwnsInput)
}
type OwnsInput struct {
	object     runtime.Object
	predicates []predicate.Predicate
}

func ControllerManagedBy(m manager.Manager) *Builder {
	return &Builder{mgr: m}
}

func (blder *Builder) loadRestConfig() {
	if blder.config == nil {
		blder.config = blder.mgr.GetConfig()
	}
}

func (blder *Builder) For(object runtime.Object, opts ...ForOption) *Builder {
	input := ForInput{object: object}
	for _, opt := range opts {
		opt.ApplyToFor(&input)
	}

	blder.forInput = input
	return blder
}

func (blder *Builder) Owns(object runtime.Object, opts ...OwnsOption) *Builder {
	input := OwnsInput{object: object}
	for _, opt := range opts {
		opt.ApplyToOwns(&input)
	}

	blder.ownsInput = append(blder.ownsInput, input)
	return blder
}

func (blder *Builder) Complete(r reconcile.Reconciler) error {
	_, err := blder.Build(r)
	return err
}

func (blder *Builder) Build(r reconcile.Reconciler) (controller.Controller, error) {
	if r == nil {
		return nil, fmt.Errorf("must provide a non-nil Reconciler")
	}
	if blder.mgr == nil {
		return nil, fmt.Errorf("must provide a non-nil Manager")
	}

	// Set the Config
	blder.loadRestConfig()

	// Set the ControllerManagedBy
	if err := blder.doController(r); err != nil {
		return nil, err
	}

	// Set the Watch
	if err := blder.doWatch(); err != nil {
		return nil, err
	}

	return blder.ctrl, nil
}
func (blder *Builder) getControllerName(gvk schema.GroupVersionKind) string {
	if blder.name != "" {
		return blder.name
	}
	return strings.ToLower(gvk.Kind)
}
func (blder *Builder) doController(reconciler reconcile.Reconciler) error {
	ctrlOptions := blder.ctrlOptions
	if ctrlOptions.Reconciler == nil {
		ctrlOptions.Reconciler = reconciler
	}

	gvk, err := apiutil.GVKForObject(blder.forInput.object, blder.mgr.GetScheme())
	if err != nil {
		return err
	}

	if ctrlOptions.Log == nil {
		ctrlOptions.Log = blder.mgr.GetLogger()
	}
	ctrlOptions.Log = ctrlOptions.Log.WithValues("reconcilerGroup", gvk.Group, "reconcilerKind", gvk.Kind)

	blder.ctrl, err = controller.New(blder.getControllerName(gvk), blder.mgr, ctrlOptions)
	return err
}

func (blder *Builder) doWatch() error {

}
