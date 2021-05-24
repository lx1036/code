package app

import (
	"context"
	"fmt"

	"k8s-lx1036/k8s/scheduler/cmd/app/config"
	"k8s-lx1036/k8s/scheduler/cmd/app/options"
	"k8s-lx1036/k8s/scheduler/pkg/scheduler"
	"k8s-lx1036/k8s/scheduler/pkg/scheduler/framework/runtime"
	"k8s-lx1036/k8s/scheduler/pkg/scheduler/profile"

	"github.com/spf13/cobra"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/client-go/tools/events"
	"k8s.io/client-go/tools/leaderelection"
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/component-base/cli/globalflag"
	"k8s.io/component-base/term"
	"k8s.io/component-base/version"
	"k8s.io/component-base/version/verflag"
	"k8s.io/klog/v2"
)

// Option configures a framework.Registry.
type Option func(runtime.Registry) error

// NewSchedulerCommand creates a *cobra.Command object with default parameters and registryOptions
func NewSchedulerCommand(registryOptions ...Option) *cobra.Command {
	opts, err := options.NewOptions()
	if err != nil {
		klog.Fatalf("unable to initialize command options: %v", err)
	}

	cmd := &cobra.Command{
		Use: "dynamic-scheduler",
		Args: func(cmd *cobra.Command, args []string) error {
			for _, arg := range args {
				if len(arg) > 0 {
					return fmt.Errorf("%q does not take any arguments, got %q", cmd.CommandPath(), args)
				}
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			if err := runCommand(cmd, opts, registryOptions...); err != nil {
				klog.Fatal(err)
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
	cmd.MarkFlagFilename("config", "yaml", "yml", "json")
	_ = cmd.MarkFlagFilename("config", "yaml", "yml", "json")

	return cmd
}

func runCommand(cmd *cobra.Command, opts *options.Options, outOfTreeRegistryOptions ...Option) error {
	verflag.PrintAndExitIfRequested()
	cliflag.PrintFlags(cmd.Flags())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cc, sched, err := Setup(ctx, opts, outOfTreeRegistryOptions...)
	if err != nil {
		return err
	}

	return Run(ctx, cc, sched)
}

func Setup(ctx context.Context, opts *options.Options,
	outOfTreeRegistryOptions ...Option) (*config.Config, *scheduler.Scheduler, error) {
	if errs := opts.Validate(); len(errs) > 0 {
		return nil, nil, utilerrors.NewAggregate(errs)
	}

	c, err := opts.Config()
	if err != nil {
		return nil, nil, err
	}

	outOfTreeRegistry := make(runtime.Registry)
	for _, option := range outOfTreeRegistryOptions {
		if err := option(outOfTreeRegistry); err != nil {
		}
	}

	// INFO: kube-scheduler 提供一种可以外部参数定义的方式，通过修改schedulerOptions进一步自定义scheduler对象，可以借鉴
	recorderFactory := getRecorderFactory(c)
	sched, err := scheduler.New(
		c.Client,
		c.InformerFactory,
		c.PodInformer,
		recorderFactory,
		ctx.Done(),
		scheduler.WithProfiles(c.ComponentConfig.Profiles...),
		scheduler.WithPercentageOfNodesToScore(c.ComponentConfig.PercentageOfNodesToScore),
		scheduler.WithFrameworkOutOfTreeRegistry(outOfTreeRegistry),
		scheduler.WithPodMaxBackoffSeconds(c.ComponentConfig.PodMaxBackoffSeconds),
		scheduler.WithPodInitialBackoffSeconds(c.ComponentConfig.PodInitialBackoffSeconds),
	)

	return c, sched, nil
}

// Run executes the scheduler based on the given configuration. It only returns on error or when context is done.
func Run(ctx context.Context, cc *config.Config, sched *scheduler.Scheduler) error {
	klog.V(1).Infof("Starting Kubernetes Scheduler version %+v", version.Get())

	// Prepare the event broadcaster
	cc.EventBroadcaster.StartRecordingToSink(ctx.Done())

	// Start all informers.
	go cc.PodInformer.Informer().Run(ctx.Done())
	cc.InformerFactory.Start(ctx.Done())

	// Wait for all caches to sync before scheduling.
	cc.InformerFactory.WaitForCacheSync(ctx.Done())

	// If leader election is enabled, runCommand via LeaderElector until done and exit.
	if cc.LeaderElection != nil {
		cc.LeaderElection.Callbacks = leaderelection.LeaderCallbacks{
			OnStartedLeading: sched.Run,
			OnStoppedLeading: func() {
				klog.Fatalf("leaderelection lost")
			},
		}
		leaderElector, err := leaderelection.NewLeaderElector(*cc.LeaderElection)
		if err != nil {
			return fmt.Errorf("couldn't create leader elector: %v", err)
		}

		leaderElector.Run(ctx)

		return fmt.Errorf("lost lease")
	}

	// Leader election is disabled, so runCommand inline until done.
	sched.Run(ctx)
	return fmt.Errorf("finished without leader elect")
}

func getRecorderFactory(cc *config.Config) profile.RecorderFactory {
	return func(name string) events.EventRecorder {
		return cc.EventBroadcaster.NewRecorder(name)
	}
}
