package app

import (
	"fmt"
	"k8s-lx1036/k8s/scheduler/descheduler/pkg/descheduler"
	"os"

	"k8s-lx1036/k8s/scheduler/descheduler/cmd/app/options"

	"github.com/spf13/cobra"

	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/component-base/cli/globalflag"
	"k8s.io/component-base/term"
	"k8s.io/component-base/version/verflag"
	"k8s.io/klog/v2"
)

// debug in local: go run .
func NewDeschedulerCommand() *cobra.Command {
	opts, err := options.NewOptions()
	if err != nil {
		klog.Fatalf("unable to initialize command options: %v", err)
	}

	cmd := &cobra.Command{
		Use:   "descheduler",
		Short: "descheduler",
		Long:  `The descheduler evicts pods which may be bound to less desired nodes`,
		Args: func(cmd *cobra.Command, args []string) error {
			for _, arg := range args {
				if len(arg) > 0 {
					return fmt.Errorf("%q does not take any arguments, got %q", cmd.CommandPath(), args)
				}
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			/*opts.Logs.LogFormat = opts.Logging.Format
			opts.Logs.Apply()

			// LoopbackClientConfig is a config for a privileged loopback connection
			var LoopbackClientConfig *restclient.Config
			var SecureServing *apiserver.SecureServingInfo
			if err := opts.SecureServing.ApplyTo(&SecureServing, &LoopbackClientConfig); err != nil {
				klog.ErrorS(err, "failed to apply secure server configuration")
				return
			}*/

			//

			if err := runCommand(cmd, opts); err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				os.Exit(1)
			}
		},
	}

	fs := cmd.Flags()
	namedFlagSets := opts.Flags()
	verflag.AddFlags(namedFlagSets.FlagSet("global"))
	globalflag.AddGlobalFlags(namedFlagSets.FlagSet("global"), cmd.Name())
	for _, f := range namedFlagSets.FlagSets {
		fs.AddFlagSet(f)
	}

	//cmd.SetOut(os.Stdout)
	//flags := cmd.Flags()
	//flags.SetNormalizeFunc(aflag.WordSepNormalizeFunc)
	//flags.AddGoFlagSet(flag.CommandLine)
	//
	//opts.AddFlags(flags)

	/*namedFlagSets := opts.Flags()
	verflag.AddFlags(namedFlagSets.FlagSet("global"))
	globalflag.AddGlobalFlags(namedFlagSets.FlagSet("global"), cmd.Name())
	for _, f := range namedFlagSets.FlagSets {
		flags.AddFlagSet(f)
	}*/

	usageFmt := "Usage:\n  %s\n"
	cols, _, _ := term.TerminalSize(cmd.OutOrStdout())
	cmd.SetUsageFunc(func(cmd *cobra.Command) error {
		fmt.Fprintf(cmd.OutOrStderr(), usageFmt, cmd.UseLine())
		cliflag.PrintSections(cmd.OutOrStderr(), namedFlagSets, cols)
		return nil
	})
	cmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		fmt.Fprintf(cmd.OutOrStdout(), "%s\n\n"+usageFmt, cmd.Long, cmd.UseLine())
		cliflag.PrintSections(cmd.OutOrStdout(), namedFlagSets, cols)
	})

	_ = cmd.MarkFlagFilename("policy-config-file", "yaml", "yml", "json")

	return cmd
}

func runCommand(cmd *cobra.Command, opts *options.Options) error {
	verflag.PrintAndExitIfRequested()
	cliflag.PrintFlags(cmd.Flags())

	klog.Infof("opts: %v", opts.KubeconfigFile)

	return descheduler.Run(opts)
}
