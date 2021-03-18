package app

import (
	"fmt"
	"io/ioutil"

	"k8s-lx1036/k8s/scheduler/pkg/scheduler/apis/config"

	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/kubernetes/pkg/scheduler"
)

// Options has all the params needed to run a Scheduler
type Options struct {

	// ConfigFile is the location of the scheduler server's configuration file.
	ConfigFile string

	Kubeconfig string
}

// Flags returns flags for a specific scheduler by section name
func (o *Options) Flags() (nfs cliflag.NamedFlagSets) {
	fs := nfs.FlagSet("misc")
	fs.StringVar(&o.ConfigFile, "config", o.ConfigFile, `The path to the configuration file.`)
	fs.StringVar(&o.Kubeconfig, "kubeconfig", o.Kubeconfig, "kubeconfig file")

	return nfs
}

// ApplyTo applies the scheduler options to the given scheduler app configuration.
func (o *Options) ApplyTo(c *Config) error {
	cfg, err := loadConfigFromFile(o.ConfigFile)
	if err != nil {
		return err
	}

	c.ComponentConfig = *cfg

	return nil
}

// Config return a scheduler config object
func (o *Options) Config() (*Config, error) {
	c := &Config{}
	if err := o.ApplyTo(c); err != nil {
		return nil, err
	}

	client, err := GetKubeClient(o)
	if err != nil {
		return nil, err
	}

	c.Client = client
	c.InformerFactory = informers.NewSharedInformerFactory(client, 0)
	c.PodInformer = scheduler.NewPodInformer(client, 0)

	return c, nil
}

// GetKubernetesClient gets the client for k8s, if ~/.kube/config exists so get that config else incluster config
func GetKubeClient(options *Options) (*kubernetes.Clientset, error) {
	c, err := clientcmd.BuildConfigFromFlags("", options.Kubeconfig)
	if err != nil {
		return nil, err
	}

	return kubernetes.NewForConfig(c)
}

func loadConfigFromFile(file string) (*config.KubeSchedulerConfiguration, error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	return loadConfig(data)
}

func loadConfig(data []byte) (*config.KubeSchedulerConfiguration, error) {
	// The UniversalDecoder runs defaulting and returns the internal type by default.
	obj, gvk, err := config.Codecs.UniversalDecoder().Decode(data, nil, nil)
	if err != nil {
		return nil, err
	}
	if cfgObj, ok := obj.(*config.KubeSchedulerConfiguration); ok {
		return cfgObj, nil
	}
	return nil, fmt.Errorf("couldn't decode as KubeSchedulerConfiguration, got %s: ", gvk)
}

// NewOptions returns default scheduler app options.
func NewOptions() (*Options, error) {
	o := &Options{}

	return o, nil
}
