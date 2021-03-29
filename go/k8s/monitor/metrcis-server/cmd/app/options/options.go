package options

import (
	"time"

	"k8s-lx1036/k8s/monitor/metrcis-server/pkg/api"
	"k8s-lx1036/k8s/monitor/metrcis-server/pkg/server"
	"k8s-lx1036/k8s/monitor/metrcis-server/pkg/version"

	"github.com/spf13/cobra"
	genericapiserver "k8s.io/apiserver/pkg/server"
)

type Options struct {
	MetricResolution time.Duration

	Kubeconfig string
}

func (o *Options) Flags(cmd *cobra.Command) {
	flags := cmd.Flags()
	flags.DurationVar(&o.MetricResolution, "metric-resolution", o.MetricResolution, "The resolution at which metrics-server will retain metrics.")
	flags.StringVar(&o.Kubeconfig, "kubeconfig", o.Kubeconfig, "The path to the kubeconfig used to connect to the Kubernetes API server and the Kubelets (defaults to in-cluster config)")

}

func (o *Options) ServerConfig() (*server.Config, error) {
	apiserverConfig := genericapiserver.NewConfig(api.Codecs)
	apiserverConfig.Version = version.VersionInfo()
	return &server.Config{
		MetricResolution: o.MetricResolution,
		ScrapeTimeout:    time.Duration(float64(o.MetricResolution) * 0.90), // scrape timeout is 90% of the scrape interval
		Kubeconfig:       o.Kubeconfig,
		ApiserverConfig:  apiserverConfig,
	}, nil
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
