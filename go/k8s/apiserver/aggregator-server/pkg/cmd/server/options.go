package server

import (
	"errors"
	"io"

	"k8s-lx1036/k8s/apiserver/aggregator-server/pkg/apiserver"
	aggregatorscheme "k8s-lx1036/k8s/apiserver/aggregator-server/pkg/apiserver/scheme"

	"github.com/spf13/pflag"

	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	genericapiserver "k8s.io/apiserver/pkg/server"
	genericoptions "k8s.io/apiserver/pkg/server/options"
)

// AggregatorOptions contains everything necessary to create and run an API Aggregator.
type AggregatorOptions struct {
	ServerRunOptions   *genericoptions.ServerRunOptions
	RecommendedOptions *genericoptions.RecommendedOptions
	APIEnablement      *genericoptions.APIEnablementOptions

	// ProxyClientCert/Key are the client cert used to identify this proxy. Backing APIServices use
	// this to confirm the proxy's identity
	ProxyClientCertFile string
	ProxyClientKeyFile  string

	StdOut io.Writer
	StdErr io.Writer
}

// AddFlags is necessary because hyperkube doesn't work using cobra, so we have to have different registration and execution paths
func (o *AggregatorOptions) AddFlags(fs *pflag.FlagSet) {
	// INFO: 这里依次调用
	o.ServerRunOptions.AddUniversalFlags(fs)
	o.RecommendedOptions.AddFlags(fs)
	o.APIEnablement.AddFlags(fs)

	fs.StringVar(&o.ProxyClientCertFile, "proxy-client-cert-file", o.ProxyClientCertFile, "client certificate used identify the proxy to the API server")
	fs.StringVar(&o.ProxyClientKeyFile, "proxy-client-key-file", o.ProxyClientKeyFile, "client certificate key used identify the proxy to the API server")
}

// Complete fills in missing Options.
func (o *AggregatorOptions) Complete() error {
	return nil
}

// Validate validates all the required options.
func (o *AggregatorOptions) Validate(args []string) error {
	errors := []error{}
	errors = append(errors, o.ServerRunOptions.Validate()...)
	errors = append(errors, o.RecommendedOptions.Validate()...)
	errors = append(errors, o.APIEnablement.Validate(aggregatorscheme.Scheme)...)
	return utilerrors.NewAggregate(errors)
}

// RunAggregator runs the API Aggregator.
func (o AggregatorOptions) RunAggregator(stopCh <-chan struct{}) error {

	serverConfig := genericapiserver.NewRecommendedConfig(aggregatorscheme.Codecs)

	if err := o.ServerRunOptions.ApplyTo(&serverConfig.Config); err != nil {
		return err
	}
	if err := o.RecommendedOptions.ApplyTo(serverConfig); err != nil {
		return err
	}
	if err := o.APIEnablement.ApplyTo(&serverConfig.Config,
		apiserver.DefaultAPIResourceConfigSource(), aggregatorscheme.Scheme); err != nil {
		return err
	}

	serviceResolver := apiserver.NewClusterIPServiceResolver(serverConfig.SharedInformerFactory.Core().V1().Services().Lister())

	config := apiserver.Config{
		GenericConfig: serverConfig,
		ExtraConfig: apiserver.ExtraConfig{
			ServiceResolver: serviceResolver,
		},
	}

	if len(o.ProxyClientCertFile) == 0 || len(o.ProxyClientKeyFile) == 0 {
		return errors.New("missing a client certificate along with a key to identify the proxy to the API server")
	}

	config.ExtraConfig.ProxyClientCertFile = o.ProxyClientCertFile
	config.ExtraConfig.ProxyClientKeyFile = o.ProxyClientKeyFile

	server, err := config.Complete().NewWithDelegate(genericapiserver.NewEmptyDelegate())
	if err != nil {
		return err
	}

	prepared, err := server.PrepareRun()
	if err != nil {
		return err
	}
	return prepared.Run(stopCh)
}
