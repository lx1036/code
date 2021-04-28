package options

import (
	"time"

	"github.com/spf13/cobra"
)

type Options struct {
	MetricResolution time.Duration

	Kubeconfig string
	Nodename   string
}

func (o *Options) Flags(cmd *cobra.Command) {
	flags := cmd.Flags()
	flags.DurationVar(&o.MetricResolution, "metric-resolution", o.MetricResolution, "The resolution at which metrics-server will retain metrics.")
	flags.StringVar(&o.Kubeconfig, "kubeconfig", o.Kubeconfig, "The path to the kubeconfig used to connect to the Kubernetes API server and the Kubelets (defaults to in-cluster config)")
	flags.StringVar(&o.Nodename, "nodename", o.Nodename, "current node name")
}

func NewOptions() *Options {
	o := &Options{
		//SecureServing:  genericoptions.NewSecureServingOptions().WithLoopback(),
		//Authentication: genericoptions.NewDelegatingAuthenticationOptions(),
		//Authorization:  genericoptions.NewDelegatingAuthorizationOptions(),
		//Features:       genericoptions.NewFeatureOptions(),
		MetricResolution: 60 * time.Second,
		//KubeletPort:                  10250,
	}

	return o
}
