package app

import (
	"k8s.io/component-base/cli/globalflag"
	"k8s.io/component-base/version/verflag"

	"github.com/spf13/cobra"

	"k8s.io/klog/v2"
)

// debug in local: go run .
func NewDeschedulerCommand() *cobra.Command {
	opts, err := NewOptions()
	if err != nil {
		klog.Fatalf("unable to initialize command options: %v", err)
	}

	cmd := &cobra.Command{
		Use:   "descheduler",
		Short: "descheduler",
		Long:  `The descheduler evicts pods which may be bound to less desired nodes`,
		Run: func(cmd *cobra.Command, args []string) {
			klog.Fatalf("opts: %v, args: %v", *opts, args)

		},
	}

	flags := cmd.Flags()
	//flags.SetNormalizeFunc(aflag.WordSepNormalizeFunc)
	//flags.AddGoFlagSet(flag.CommandLine)

	//opts.AddFlags(flags)

	namedFlagSets := opts.Flags()
	verflag.AddFlags(namedFlagSets.FlagSet("global"))
	globalflag.AddGlobalFlags(namedFlagSets.FlagSet("global"), cmd.Name())
	for _, f := range namedFlagSets.FlagSets {
		flags.AddFlagSet(f)
	}

	return cmd
}
