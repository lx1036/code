/*

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	k8slx1036comv1 "k8s-lx1036/api/v1"
)

// CustomGrafanaDashboardReconciler reconciles a CustomGrafanaDashboard object
type CustomGrafanaDashboardReconciler struct {
	client.Client
	//Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=k8s.lx1036.com.k8s.lx1036.com,resources=customgrafanadashboards,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=k8s.lx1036.com.k8s.lx1036.com,resources=customgrafanadashboards/status,verbs=get;update;patch

func (r *CustomGrafanaDashboardReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	_ = r.Log.WithValues("customgrafanadashboard", req.NamespacedName)

	// your logic here

	return ctrl.Result{}, nil
}

func (r *CustomGrafanaDashboardReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&k8slx1036comv1.CustomGrafanaDashboard{}).
		Complete(r)
}
