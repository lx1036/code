package app

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"io/ioutil"
	v1 "k8s-lx1036/k8s/kubelet/cri-hook-server/pkg/apis/crihookserver.k9s.io/v1"
	"k8s-lx1036/k8s/kubelet/cri-hook-server/pkg/server"
	genericapiserver "k8s.io/apiserver/pkg/server"
	
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
)




func NewCriHookServerCommand() *cobra.Command {
	opts := NewOptions()
	
	cmd := &cobra.Command{
		Use: "lighthouse",
		Long: "The lighthouse runs on each node. This is a preHook framework to modify request body for " +
			"any matched rules in the configuration. It is an enhancement for kubelet to run a container",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Flags().VisitAll(func(flag *pflag.Flag) {
				klog.Infof("FLAG: --%s=%q", flag.Name, flag.Value)
			})
			
			if err := opts.Complete(); err != nil {
				klog.Fatalf("failed complete: %v", err)
			}
			
			if err := opts.Run(); err != nil {
				klog.Exit(err)
			}
		},
	}
	
	opts.AddFlags(cmd.Flags())
	
	return cmd
}


type Options struct {
	ConfigFile string
	config     *v1.HookConfiguration
}


func NewOptions() *Options {
	return &Options{
		config: &v1.HookConfiguration{},
	}
}

func (o *Options) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.ConfigFile, "config", o.ConfigFile, "The path to the configuration file")
}

// Complete INFO: 把 config file decode into v1.HookConfiguration
func (o *Options) Complete() error {
	if len(o.ConfigFile) > 0 {
		cfgData, err := ioutil.ReadFile(o.ConfigFile)
		if err != nil {
			return fmt.Errorf("failed to read hook configuration file %q, %v", o.ConfigFile, err)
		}
		
		// decode hook configuration
		versioned := &v1.HookConfiguration{}
		v1.SetObjectDefaults_HookConfiguration(versioned)
		decoder := v1.Codecs.UniversalDecoder(v1.SchemeGroupVersion)
		if err := runtime.DecodeInto(decoder, cfgData, versioned); err != nil {
			return fmt.Errorf("failed to decode hook configuration file %q, %v", o.ConfigFile, err)
		}
		
		// convert versioned hook configuration to internal version
		if err := v1.Scheme.Convert(versioned, o.config, nil); err != nil {
			return fmt.Errorf("failed to convert versioned hook configurtion to internal version, %v", err)
		}
		
		return nil
	}
	
	return fmt.Errorf("config file is required")
}

func (o *Options) Run() error {
	hookServer := server.NewHookServer(o.config)
	
	return hookServer.Run(genericapiserver.SetupSignalHandler())
}
