package main

import (
	"flag"
	v1 "k8s-lx1036/k8s/concepts/kubebuilder/controller-runtime/lvs/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"os"
	controllers "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"time"
)

// https://github.com/kubernetes-sigs/controller-runtime/blob/master/example_test.go

var (
	scheme   = runtime.NewScheme()
	setupLog = controllers.Log.WithName("setup")
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)
	_ = v1.AddToScheme(scheme)
}

// go run . --kubeconfig=/Users/liuxiang/.kube/config
func main() {
	flag.Parse()
	controllers.SetLogger(zap.New(func(o *zap.Options) {
		o.Development = true
	}))
	
	leaseDuration := 100 * time.Second
	renewDeadline := 80 * time.Second
	retryPeriod := 20 * time.Second
	
	mgr, err := controllers.NewManager(controllers.GetConfigOrDie(), controllers.Options{
		LeaseDuration: &leaseDuration,
		RenewDeadline: &renewDeadline,
		RetryPeriod:   &retryPeriod,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}
	
	//Replicaset(mgr)
	
	LvsDeployment(mgr)
	
	if err := mgr.Start(controllers.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "could not start manager")
		os.Exit(1)
	}
}
