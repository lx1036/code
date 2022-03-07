package app

import (
	"k8s-lx1036/k8s/network/loadbalancer/bgplb/cmd/app/options"
	"k8s-lx1036/k8s/network/loadbalancer/bgplb/pkg/controller"

	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
)

func NewBGPLBCommand(stopCh <-chan struct{}) *cobra.Command {
	opts := options.NewOptions()

	cmd := &cobra.Command{
		Short: "Launch bgplb",
		Long:  "Launch bgplb",
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

func runCommand(option *options.Options, stopCh <-chan struct{}) error {
	bgpLBController, err := controller.NewController(option)
	if err != nil {
		return err
	}
	klog.Info("starting run server...")
	err = bgpLBController.Start(option.ControllerThreads, stopCh)
	if err != nil {
		return err
	}

	<-stopCh

	// INFO: 这里做了queue清理工作，可以借鉴下，不过不是很重要
	klog.Info("Shutting down the etcd cluster")
	bgpLBController.Stop()
	return nil
}
