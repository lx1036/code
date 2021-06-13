package options

import (
	"github.com/spf13/cobra"
	apiv1 "k8s.io/api/core/v1"
)

type Options struct {
	// path to the kubeconfig used to connect to the Kubernetes API server
	Kubeconfig string

	vpaObjectNamespace string

	ControllerThreads int
}

func (o *Options) Flags(cmd *cobra.Command) {
	flags := cmd.Flags()
	flags.StringVar(&o.Kubeconfig, "kubeconfig", o.Kubeconfig, "The path to the kubeconfig used to connect to the Kubernetes API server and the Kubelets (defaults to in-cluster config)")
	flags.StringVar(&o.vpaObjectNamespace, "vpa-object-namespace", apiv1.NamespaceAll, "Namespace to search for VPA objects and pod stats. Empty means all namespaces will be used.")
	flags.IntVar(&o.ControllerThreads, "controller-threads", 1, "Number of worker threads used by the SparkApplication controller.")
}

func NewOptions() *Options {
	return &Options{}
}
