package options

import (
	"time"

	"github.com/spf13/cobra"
)

const (
	defaultSchedulerName   = "kube-batch"
	defaultSchedulerPeriod = time.Second
	defaultQueue           = "default"
	defaultListenAddress   = ":8080"
)

type Options struct {
	Master               string
	Kubeconfig           string
	SchedulerName        string
	SchedulerConf        string
	SchedulePeriod       time.Duration
	EnableLeaderElection bool
	LockObjectNamespace  string
	DefaultQueue         string
	PrintVersion         bool
	ListenAddress        string
	EnablePriorityClass  bool
}

func (o *Options) Flags(cmd *cobra.Command) {
	flags := cmd.Flags()
	flags.StringVar(&o.Kubeconfig, "kubeconfig", o.Kubeconfig, "The path to the kubeconfig used to connect to the Kubernetes API server and the Kubelets (defaults to in-cluster config)")
	flags.StringVar(&o.Master, "master", o.Master, "The address of the Kubernetes API server (overrides any value in kubeconfig)")

	flags.StringVar(&o.SchedulerName, "scheduler-name", defaultSchedulerName, "kube-batch will handle pods whose .spec.SchedulerName is same as scheduler-name")
	flags.StringVar(&o.SchedulerConf, "scheduler-conf", "", "The absolute path of scheduler configuration file")
	flags.DurationVar(&o.SchedulePeriod, "schedule-period", defaultSchedulerPeriod, "The period between each scheduling cycle")
	flags.StringVar(&o.DefaultQueue, "default-queue", defaultQueue, "The default queue name of the job")

	flags.BoolVar(&o.EnableLeaderElection, "leader-elect", o.EnableLeaderElection,
		"Start a leader election client and gain leadership before "+
			"executing the main loop. Enable this when running replicated kube-batch for high availability")
	flags.BoolVar(&o.PrintVersion, "version", false, "Show version and quit")
	flags.StringVar(&o.LockObjectNamespace, "lock-object-namespace", o.LockObjectNamespace, "Define the namespace of the lock object that is used for leader election")

	flags.StringVar(&o.ListenAddress, "listen-address", defaultListenAddress, "The address to listen on for HTTP requests.")
	flags.BoolVar(&o.EnablePriorityClass, "priority-class", true,
		"Enable PriorityClass to provide the capacity of preemption at pod group level; to disable it, set it false")
}

func NewOptions() *Options {
	o := &Options{}

	return o
}
