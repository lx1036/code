package server

import (
	"io"

	"k8s-lx1036/k8s/apiserver/aggregator-server/pkg/apis/apiregistration/v1"
	aggregatorscheme "k8s-lx1036/k8s/apiserver/aggregator-server/pkg/apiserver/scheme"

	"github.com/spf13/cobra"

	genericoptions "k8s.io/apiserver/pkg/server/options"
)

const defaultEtcdPathPrefix = "/registry/kube-aggregator.kubernetes.io/"

// NewDefaultOptions builds a "normal" set of options.  You wouldn't normally expose this, but hyperkube isn't cobra compatible
func NewDefaultOptions(out, err io.Writer) *AggregatorOptions {
	o := &AggregatorOptions{
		ServerRunOptions: genericoptions.NewServerRunOptions(),
		RecommendedOptions: genericoptions.NewRecommendedOptions(
			defaultEtcdPathPrefix,
			aggregatorscheme.Codecs.LegacyCodec(v1.SchemeGroupVersion),
		),
		APIEnablement: genericoptions.NewAPIEnablementOptions(),

		StdOut: out,
		StdErr: err,
	}

	return o
}

// NewCommandStartAggregator provides a CLI handler for 'start master' command
// with a default AggregatorOptions.
func NewCommandStartAggregator(defaults *AggregatorOptions, stopCh <-chan struct{}) *cobra.Command {
	o := *defaults
	cmd := &cobra.Command{
		Short: "Launch a API aggregator and proxy server",
		Long:  "Launch a API aggregator and proxy server",
		RunE: func(c *cobra.Command, args []string) error {
			if err := o.Complete(); err != nil {
				return err
			}
			if err := o.Validate(args); err != nil {
				return err
			}
			if err := o.RunAggregator(stopCh); err != nil {
				return err
			}
			return nil
		},
	}

	o.AddFlags(cmd.Flags())

	return cmd
}
