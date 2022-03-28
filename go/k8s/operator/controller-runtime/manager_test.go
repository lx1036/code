package controller_runtime

import (
	corev1 "k8s.io/api/core/v1"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
	"testing"
)

var (
	mgr manager.Manager
	// NB: don't call SetLogger in init(), or else you'll mess up logging in the main suite.
	log = logf.Log.WithName("manager-examples")
)

func TestManager(test *testing.T) {
	cfg := config.GetConfigOrDie()
	mgr, err := manager.New(cfg, manager.Options{
		NewCache: cache.MultiNamespacedCacheBuilder([]string{"default"}),
	})
	if err != nil {
		log.Error(err, "unable to set up manager")
		os.Exit(1)
	}
	log.Info("created manager", "manager", mgr)

	ns := corev1.Namespace{}
	recorder := mgr.GetEventRecorderFor("test-manager")
	err = mgr.Add(manager.RunnableFunc(func(stop <-chan struct{}) error {
		// Do something
		recorder.Event(&ns, "Warning", "reason", "message")

		return nil
	}))
	if err != nil {
		log.Error(err, "unable add a runnable to the manager")
		os.Exit(1)
	}

	err = mgr.Start(signals.SetupSignalHandler())
	if err != nil {
		log.Error(err, "unable start the manager")
		os.Exit(1)
	}
}
