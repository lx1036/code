package main

import (
	"context"
	"github.com/go-logr/logr"
	v1 "k8s-lx1036/k8s/concepts/kubebuilder/controller-runtime/lvs/v1"
	appsv1 "k8s.io/api/apps/v1"
	"os"
	controllers "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"strings"
)

type DeploymentReconciler struct {
	client.Client
	Log logr.Logger
}

func (r DeploymentReconciler) Reconcile(request controllers.Request) (controllers.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("deployment", request.NamespacedName)
	var lvsPodList v1.LvsPodList
	if err := r.List(ctx, &lvsPodList, client.InNamespace(request.Namespace)); err != nil {
		log.Error(err, "unable list LvsPod")
		return controllers.Result{}, err
	}
	if len(lvsPodList.Items) != 0 {
		var names []string
		for _, lvsPod := range lvsPodList.Items {
			names = append(names, lvsPod.Name)
		}

		log.V(1).Info("list LvsPod", "names", strings.Join(names, ","))
	}

	return controllers.Result{}, nil
}

func LvsDeployment(manager manager.Manager) {
	err := controllers.
		NewControllerManagedBy(manager). // Create the Controller
		For(&appsv1.Deployment{}).       // ReplicaSet is the Application API
		Complete(&DeploymentReconciler{
			Client: manager.GetClient(),
			Log:    controllers.Log.WithName("controllers").WithName("LvsPod"),
		})
	if err != nil {
		setupLog.Error(err, "could not create controller")
		os.Exit(1)
	}
}
