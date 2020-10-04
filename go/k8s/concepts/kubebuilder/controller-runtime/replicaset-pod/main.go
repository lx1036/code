package main

import (
	"context"
	"flag"
	"fmt"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"os"
	controllers "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"time"
)

// https://github.com/kubernetes-sigs/controller-runtime/blob/master/example_test.go

var (
	scheme   = runtime.NewScheme()
	setupLog = controllers.Log.WithName("setup")
)

// go run . --kubeconfig=/Users/liuxiang/.kube/config
func main() {
	Replicaset()
}

func Replicaset() {
	flag.Parse()
	controllers.SetLogger(zap.New(func(o *zap.Options) {
		o.Development = true
	}))

	leaseDuration := 100 * time.Second
	renewDeadline := 80 * time.Second
	retryPeriod := 20 * time.Second

	manager, err := controllers.NewManager(controllers.GetConfigOrDie(), controllers.Options{
		LeaseDuration: &leaseDuration,
		RenewDeadline: &renewDeadline,
		RetryPeriod:   &retryPeriod,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	err = controllers.
		NewControllerManagedBy(manager). // Create the Controller
		For(&appsv1.ReplicaSet{}).       // ReplicaSet is the Application API
		Owns(&corev1.Pod{}).             // ReplicaSet owns Pods created by it
		Complete(&ReplicaSetReconciler{Client: manager.GetClient()})
	if err != nil {
		setupLog.Error(err, "could not create controller")
		os.Exit(1)
	}

	if err := manager.Start(controllers.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "could not start manager")
		os.Exit(1)
	}
}

type ReplicaSetReconciler struct {
	client.Client
}

// Implement the business logic:
// This function will be called when there is a change to a ReplicaSet or a Pod with an OwnerReference
// to a ReplicaSet.
//
// * Read the ReplicaSet
// * Read the Pods
// * Set a Label on the ReplicaSet with the Pod count
func (reconciler *ReplicaSetReconciler) Reconcile(request controllers.Request) (controllers.Result, error) {
	// Read the ReplicaSet
	rs := &appsv1.ReplicaSet{}
	ctx := context.TODO()
	err := reconciler.Get(ctx, request.NamespacedName, rs)
	if err != nil {
		return controllers.Result{}, err
	}

	// List the Pods matching the PodTemplate Labels
	pods := &corev1.PodList{}
	err = reconciler.List(ctx, pods, client.InNamespace(request.Namespace), client.MatchingLabels(rs.Spec.Template.Labels))
	if err != nil {
		return controllers.Result{}, err
	}

	// Update the ReplicaSet
	rs.Labels["pod-count"] = fmt.Sprintf("%v", len(pods.Items))
	err = reconciler.Update(context.TODO(), rs)
	if err != nil {
		return controllers.Result{}, err
	}

	return controllers.Result{}, nil
}
