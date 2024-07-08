package app

import (
	"k8s-lx1036/k8s/scheduler/volcano/volcano/cmd/scheduler/app/options"
	"k8s-lx1036/k8s/scheduler/volcano/volcano/pkg/kube"
	"k8s-lx1036/k8s/scheduler/volcano/volcano/pkg/scheduler"
	"k8s-lx1036/k8s/scheduler/volcano/volcano/pkg/version"
	"k8s.io/klog/v2"
	"net/http"
)

func Run(opt *options.ServerOption) error {
	version.PrintVersion()

	config, err := kube.BuildConfig(opt.KubeClientOptions)
	if err != nil {
		return err
	}

	sched, err := scheduler.NewScheduler(config, opt)
	if err != nil {
		klog.Fatalf("new scheduler err: %v", err)
		return err
	}

	if opt.EnableMetrics {
		go func() {
			http.Handle("/metrics", promHandler())
			klog.Fatalf("Prometheus Http Server failed %s", http.ListenAndServe(opt.ListenAddress, nil))
		}()
	}

}
