package app

import (
	"github.com/spf13/pflag"
	clientset "k8s.io/client-go/kubernetes"
	cliflag "k8s.io/component-base/cli/flag"

	"k8s-lx1036/k8s/scheduler/descheduler/pkg/apis/componentconfig"
	componentconfigv1alpha1 "k8s-lx1036/k8s/scheduler/descheduler/pkg/apis/componentconfig/v1alpha1"
	deschedulerscheme "k8s-lx1036/k8s/scheduler/descheduler/pkg/scheme"
	//apiserveroptions "k8s.io/apiserver/pkg/server/options"
)

const (
	DefaultDeschedulerPort = 10258
)

type Options struct {
	Client clientset.Interface

	componentconfig.DeschedulerConfiguration

	//SecureServing  *apiserveroptions.SecureServingOptionsWithLoopback

}

func (o *Options) AddFlags(fs *pflag.FlagSet) {
	//fs := nfs.FlagSet("misc")
	fs.StringVar(&o.KubeconfigFile, "kubeconfig", o.KubeconfigFile, `The path to the configuration file.`)
	fs.DurationVar(&o.DeschedulingInterval, "descheduling-interval", o.DeschedulingInterval,
		"Time interval between two consecutive descheduler executions. Setting this value instructs the descheduler to run in a continuous loop at the interval specified.")

	//o.SecureServing.AddFlags(fs)

	//return nfs
}

func (o *Options) Flags() (nfs cliflag.NamedFlagSets) {
	fs := nfs.FlagSet("misc")
	fs.StringVar(&o.KubeconfigFile, "kubeconfig", o.KubeconfigFile, `The path to the configuration file.`)

	return nfs
}

func newDefaultComponentConfig() (*componentconfig.DeschedulerConfiguration, error) {
	versionedCfg := componentconfigv1alpha1.DeschedulerConfiguration{}
	deschedulerscheme.Scheme.Default(&versionedCfg)
	cfg := componentconfig.DeschedulerConfiguration{}
	// componentconfigv1alpha1 转换版本到内部版本
	if err := deschedulerscheme.Scheme.Convert(&versionedCfg, &cfg, nil); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func NewOptions() (*Options, error) {
	config, err := newDefaultComponentConfig()
	if err != nil {
		return nil, err
	}

	//secureServing := apiserveroptions.NewSecureServingOptions().WithLoopback()
	//secureServing.BindPort = DefaultDeschedulerPort

	o := &Options{
		DeschedulerConfiguration: *config,
		//SecureServing:            secureServing,
	}

	return o, nil
}
