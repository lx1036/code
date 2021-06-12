package options

import (
	"github.com/spf13/cobra"
	apiv1 "k8s.io/api/core/v1"
)

type Options struct {
	// path to the kubeconfig used to connect to the Kubernetes API server
	Kubeconfig string

	Debug bool

	Namespace string

	ControllerThreads int
}

func (o *Options) Flags(cmd *cobra.Command) {
	flags := cmd.Flags()
	flags.BoolVar(&o.Debug, "debug", false, "debug for skip updating container cpuset")
	flags.StringVar(&o.Kubeconfig, "kubeconfig", o.Kubeconfig, "The path to the kubeconfig used to connect to the Kubernetes API server and the Kubelets (defaults to in-cluster config)")
	flags.StringVar(&o.Namespace, "namespace", apiv1.NamespaceAll, "The Kubernetes namespace to manage. Will manage custom resource objects of the managed CRD types for the whole cluster if unset.")
	flags.IntVar(&o.ControllerThreads, "controller-threads", 1, "Number of worker threads used by the SparkApplication controller.")

}

func NewOptions() *Options {
	return &Options{}
}
