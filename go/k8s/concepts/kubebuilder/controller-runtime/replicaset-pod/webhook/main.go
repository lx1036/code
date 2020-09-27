package main

import (
	"context"
	"fmt"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	controllers "sigs.k8s.io/controller-runtime"
	"os"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	setupLog = controllers.Log.WithName("setup")
)

// go run . --kubeconfig=/Users/liuxiang/.kube/config
func main() {
	run()
}

func run() {
	// Setup a Manager
	setupLog.Info("setting up manager")
	mgr, err := manager.New(config.GetConfigOrDie(), manager.Options{})
	if err != nil {
		setupLog.Error(err, "unable to set up overall controller manager")
		os.Exit(1)
	}
	
	// Setup a new controller to reconcile ReplicaSets
	setupLog.Info("Setting up controller")
	c, err := controller.New("foo-controller", mgr, controller.Options{
		Reconciler: &reconcileReplicaSet{client: mgr.GetClient()},
	})
	if err != nil {
		setupLog.Error(err, "unable to set up individual controller")
		os.Exit(1)
	}
	
	// Watch ReplicaSets and enqueue ReplicaSet object key
	if err := c.Watch(&source.Kind{Type: &appsv1.ReplicaSet{}}, &handler.EnqueueRequestForObject{}); err != nil {
		setupLog.Error(err, "unable to watch ReplicaSets")
		os.Exit(1)
	}
	
	// Watch Pods and enqueue owning ReplicaSet key
	if err := c.Watch(&source.Kind{Type: &corev1.Pod{}},
		&handler.EnqueueRequestForOwner{OwnerType: &appsv1.ReplicaSet{}, IsController: true}); err != nil {
		setupLog.Error(err, "unable to watch Pods")
		os.Exit(1)
	}
	
	// Setup webhooks
	setupLog.Info("setting up webhook server")
	hookServer := mgr.GetWebhookServer()
	
	setupLog.Info("registering webhooks to the webhook server")
	hookServer.Register("/mutate-v1-pod", &webhook.Admission{Handler: &podAnnotator{Client: mgr.GetClient()}})
	hookServer.Register("/validate-v1-pod", &webhook.Admission{Handler: &podValidator{Client: mgr.GetClient()}})
	
	setupLog.Info("starting manager")
	if err := mgr.Start(controllers.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "unable to run manager")
		os.Exit(1)
	}
}


// reconcileReplicaSet reconciles ReplicaSets
type reconcileReplicaSet struct {
	// client can be used to retrieve objects from the APIServer.
	client client.Client
}

// Implement reconcile.Reconciler so the controller can reconcile objects
var _ reconcile.Reconciler = &reconcileReplicaSet{}

func (r *reconcileReplicaSet) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	// set up a convenient log object so we don't have to type request over and over again
	
	// Fetch the ReplicaSet from the cache
	rs := &appsv1.ReplicaSet{}
	err := r.client.Get(context.TODO(), request.NamespacedName, rs)
	if errors.IsNotFound(err) {
		setupLog.Error(nil, "Could not find ReplicaSet")
		return reconcile.Result{}, nil
	}
	
	if err != nil {
		return reconcile.Result{}, fmt.Errorf("could not fetch ReplicaSet: %+v", err)
	}
	
	// Print the ReplicaSet
	setupLog.Info("Reconciling ReplicaSet", "container name", rs.Spec.Template.Spec.Containers[0].Name)
	
	// Set the label if it is missing
	if rs.Labels == nil {
		rs.Labels = map[string]string{}
	}
	if rs.Labels["hello"] == "world" {
		return reconcile.Result{}, nil
	}
	
	// Update the ReplicaSet
	rs.Labels["hello"] = "world"
	err = r.client.Update(context.TODO(), rs)
	if err != nil {
		return reconcile.Result{}, fmt.Errorf("could not write ReplicaSet: %+v", err)
	}
	
	return reconcile.Result{}, nil
}
