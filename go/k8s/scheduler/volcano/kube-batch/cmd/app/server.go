package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"k8s-lx1036/k8s/scheduler/volcano/kube-batch/cmd/app/options"
	"k8s-lx1036/k8s/scheduler/volcano/kube-batch/pkg/scheduler"
	"k8s-lx1036/k8s/scheduler/volcano/kube-batch/pkg/version"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/uuid"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
)

const (
	leaseDuration = 15 * time.Second
	renewDeadline = 10 * time.Second
	retryPeriod   = 5 * time.Second
	apiVersion    = "v1"
)

func NewKubeBatchCommand(stopCh <-chan struct{}) *cobra.Command {
	opts := options.NewOptions()

	cmd := &cobra.Command{
		Short: "Launch kube-batch",
		Long:  "Launch kube-batch",
		RunE: func(c *cobra.Command, args []string) error {
			if err := runCommand(opts, stopCh); err != nil {
				return err
			}
			return nil
		},
	}
	opts.Flags(cmd)
	return cmd
}

func runCommand(opt *options.Options, stopCh <-chan struct{}) error {
	if opt.PrintVersion {
		version.PrintVersionAndExit(apiVersion)
	}

	config, err := buildConfig(opt.Master, opt.Kubeconfig)
	if err != nil {
		return err
	}

	sched, err := scheduler.NewScheduler(config,
		opt.SchedulerName,
		opt.SchedulerConf,
		opt.SchedulePeriod,
		opt.DefaultQueue)
	if err != nil {
		panic(err)
	}

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		klog.Fatalf("Prometheus Http Server failed %s", http.ListenAndServe(opt.ListenAddress, nil))
	}()

	run := func(ctx context.Context) {
		sched.Run(ctx.Done())
		<-ctx.Done()
	}

	if !opt.EnableLeaderElection {
		run(context.TODO())
		return fmt.Errorf("finished without leader elect")
	}

	// INFO: leader election, 可以复用!!!

	leaderElectionClient, err := clientset.NewForConfig(restclient.AddUserAgent(config, "leader-election"))
	if err != nil {
		return err
	}

	// Prepare event clients.
	broadcaster := record.NewBroadcaster()
	broadcaster.StartRecordingToSink(&corev1.EventSinkImpl{Interface: leaderElectionClient.CoreV1().Events(opt.LockObjectNamespace)})
	eventRecorder := broadcaster.NewRecorder(scheme.Scheme, v1.EventSource{Component: opt.SchedulerName})
	hostname, err := os.Hostname()
	if err != nil {
		return fmt.Errorf("unable to get hostname: %v", err)
	}
	// add a uniquifier so that two processes on the same host don't accidentally both become active
	id := hostname + "_" + string(uuid.NewUUID())
	rl, err := resourcelock.New(resourcelock.LeasesResourceLock,
		opt.LockObjectNamespace,
		"kube-batch",
		leaderElectionClient.CoreV1(),
		leaderElectionClient.CoordinationV1(),
		resourcelock.ResourceLockConfig{
			Identity:      id,
			EventRecorder: eventRecorder,
		})
	if err != nil {
		return fmt.Errorf("couldn't create resource lock: %v", err)
	}
	leaderelection.RunOrDie(context.TODO(), leaderelection.LeaderElectionConfig{
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

func buildConfig(master, kubeconfig string) (*rest.Config, error) {
	if master != "" || kubeconfig != "" {
		return clientcmd.BuildConfigFromFlags(master, kubeconfig)
	}
	return rest.InClusterConfig()
}
