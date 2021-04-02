package app

import cliflag "k8s.io/component-base/cli/flag"

// Options has all the params needed to run a Scheduler
type Options struct {
	Kubeconfig string
}

func (o *Options) Flags() (nfs cliflag.NamedFlagSets) {
	fs := nfs.FlagSet("misc")
	fs.StringVar(&o.Kubeconfig, "kubeconfig", o.Kubeconfig, "kubeconfig file")

	return nfs
}

func NewOptions() (*Options, error) {
	o := &Options{}

	return o, nil
}
