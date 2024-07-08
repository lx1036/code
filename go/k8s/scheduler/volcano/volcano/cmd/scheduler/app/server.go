package app

import (
	"context"
	"fmt"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	"net/http"
	"os"
	"time"

	"k8s-lx1036/k8s/scheduler/volcano/volcano/cmd/scheduler/app/options"
	"k8s-lx1036/k8s/scheduler/volcano/volcano/pkg/kube"
	"k8s-lx1036/k8s/scheduler/volcano/volcano/pkg/scheduler"
	"k8s-lx1036/k8s/scheduler/volcano/volcano/pkg/signals"
	"k8s-lx1036/k8s/scheduler/volcano/volcano/pkg/util"
	"k8s-lx1036/k8s/scheduler/volcano/volcano/pkg/version"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/uuid"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
)

const (
	leaseDuration = 15 * time.Second
	renewDeadline = 10 * time.Second
	retryPeriod   = 5 * time.Second
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

	ctx := signals.SetupSignalContext()
	run := func(ctx context.Context) {
		sched.Run(ctx.Done())
		<-ctx.Done()
	}

	if !opt.EnableLeaderElection {
		run(ctx)
		return fmt.Errorf("finished without leader elect")
	}

	leaderElectionClient, err := clientset.NewForConfig(restclient.AddUserAgent(config, "leader-election"))
	if err != nil {
		return err
	}

	// Prepare event clients.
	broadcaster := record.NewBroadcaster()
	broadcaster.StartRecordingToSink(&corev1.EventSinkImpl{Interface: leaderElectionClient.CoreV1().Events(opt.LockObjectNamespace)})
	eventRecorder := broadcaster.NewRecorder(scheme.Scheme, v1.EventSource{Component: util.GenerateComponentName(opt.SchedulerNames)})
	hostname, err := os.Hostname()
	if err != nil {
		return fmt.Errorf("unable to get hostname: %v", err)
	}
	// add an uniquifier so that two processes on the same host don't accidentally both become active
	id := hostname + "_" + string(uuid.NewUUID())
	rl, err := resourcelock.New(resourcelock.LeasesResourceLock,
		opt.LockObjectNamespace,
		util.GenerateComponentName(opt.SchedulerNames),
		leaderElectionClient.CoreV1(),
		leaderElectionClient.CoordinationV1(),
		resourcelock.ResourceLockConfig{
			Identity:      id,
			EventRecorder: eventRecorder,
		})
	if err != nil {
		return fmt.Errorf("couldn't create resource lock: %v", err)
	}
	leaderelection.RunOrDie(ctx, leaderelection.LeaderElectionConfig{
		Lock:          rl,
		LeaseDuration: leaseDuration,
		RenewDeadline: renewDeadline,
		RetryPeriod:   retryPeriod,
		Callbacks: leaderelection.LeaderCallbacks{
			OnStartedLeading: run,
			OnStoppedLeading: func() {
				klog.Fatalf("leaderelection lost")
			},
		},
	})

	return fmt.Errorf("lost lease")
}
