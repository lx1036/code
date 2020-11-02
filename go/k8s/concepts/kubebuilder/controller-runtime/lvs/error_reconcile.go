package main

import (
	"errors"
	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	"os"
	controllers "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"time"
)

type DeploymentErrorReconciler struct {
	client.Client
	Log logr.Logger
}

func (r DeploymentErrorReconciler) Reconcile(request controllers.Request) (controllers.Result, error) {
	log := r.Log.WithValues("deployment", request.NamespacedName)
	log.V(1).Info("failed at: ", "time at: ", time.Now().Format(time.RFC3339))
	return controllers.Result{}, errors.New("err")
}

func LvsDeploymentError(manager manager.Manager) {
	err := controllers.
		NewControllerManagedBy(manager). // Create the Controller
		For(&appsv1.Deployment{}).       // ReplicaSet is the Application API
		Complete(&DeploymentErrorReconciler{
			Client: manager.GetClient(),
			Log:    controllers.Log.WithName("controllers").WithName("DeploymentError"),
		})
	if err != nil {
		setupLog.Error(err, "could not create controller")
		os.Exit(1)
	}
}
