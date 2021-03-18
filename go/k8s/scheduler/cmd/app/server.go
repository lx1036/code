package app

import (
	"context"
	"fmt"

	"k8s-lx1036/k8s/scheduler/pkg/scheduler"
	"k8s-lx1036/k8s/scheduler/pkg/scheduler/framework/runtime"

	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
)

// Option configures a framework.Registry.
type Option func(runtime.Registry) error

// NewSchedulerCommand creates a *cobra.Command object with default parameters and registryOptions
func NewSchedulerCommand(registryOptions ...Option) *cobra.Command {
	opts, err := NewOptions()
	if err != nil {
		klog.Fatalf("unable to initialize command options: %v", err)
	}

	cmd := &cobra.Command{
		Use: "scheduler",
		Args: func(cmd *cobra.Command, args []string) error {
			for _, arg := range args {
				if len(arg) > 0 {
					return fmt.Errorf("%q does not take any arguments, got %q", cmd.CommandPath(), args)
				}
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			klog.Fatalf("opts: %v", *opts)

			if err := runCommand(cmd, opts, registryOptions...); err != nil {
				klog.Fatal(err)
			}
		},
	}

	fs := cmd.Flags()
	namedFlagSets := opts.Flags()
	for _, f := range namedFlagSets.FlagSets {
		fs.AddFlagSet(f)
	}

	_ = cmd.MarkFlagFilename("config", "yaml", "yml", "json")

	return cmd
}

func runCommand(cmd *cobra.Command, opts *Options, outOfTreeRegistryOptions ...Option) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c, err := opts.Config()
	if err != nil {
		return err
	}

	outOfTreeRegistry := make(runtime.Registry)
	for _, option := range outOfTreeRegistryOptions {
		if err := option(outOfTreeRegistry); err != nil {
			return err
		}
	}

	sched, err := scheduler.New(
		c.Client,
		c.InformerFactory,
		c.PodInformer,
		ctx.Done(),
	)
	if err != nil {
		return err
	}

	// Leader election is disabled, so runCommand inline until done.
	sched.Run(ctx)
	return fmt.Errorf("finished without leader elect")
}
