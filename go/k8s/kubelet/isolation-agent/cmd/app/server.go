package app

import (
	"k8s-lx1036/k8s/kubelet/isolation-agent/cmd/app/options"
	"k8s-lx1036/k8s/kubelet/isolation-agent/pkg/server"

	"github.com/spf13/cobra"
)

func NewIsolationCommand(stopCh <-chan struct{}) *cobra.Command {
	opts := options.NewOptions()

	cmd := &cobra.Command{
		Use:  "isolation",
		Long: "isolation agent for container cpu/memory",
		RunE: func(c *cobra.Command, args []string) error {
			if err := runCommand(opts, stopCh); err != nil {
				return err
			}
			return nil
		},
	}
	opts.Flags(cmd)
	return cmd
}

func runCommand(o *options.Options, stopCh <-chan struct{}) error {
	config, err := o.ServerConfig()
	if err != nil {
		return err
	}

	s, err := server.NewServer(config)
	if err != nil {
		return err
	}

	return s.RunUntil(stopCh)
}
