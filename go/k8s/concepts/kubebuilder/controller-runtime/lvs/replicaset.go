package main

import (
	"context"
	"fmt"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	controllers "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"os"
)

type ReplicaSetReconciler struct {
	client.Client
}

// 实现逻辑：
// kubectl get rs {name} -n {namespace} // 找出ReplicaSet对象
// kubectl get pods -l app=nginx-demo,pod-template-hash=5c48484d9d 找出pods
// update/patch ReplicaSet对象的labels
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

func Replicaset(manager manager.Manager) {
	err := controllers.
		NewControllerManagedBy(manager). // Create the Controller
		For(&appsv1.ReplicaSet{}).       // ReplicaSet is the Application API
		Owns(&corev1.Pod{}).             // ReplicaSet owns Pods created by it
		Complete(&ReplicaSetReconciler{Client: manager.GetClient()})
	if err != nil {
		setupLog.Error(err, "could not create controller")
		os.Exit(1)
	}
}
