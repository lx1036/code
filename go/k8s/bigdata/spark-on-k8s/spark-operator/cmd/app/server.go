package app

import (
	"k8s-lx1036/k8s/bigdata/spark-on-k8s/spark-operator/cmd/app/options"
	"k8s-lx1036/k8s/bigdata/spark-on-k8s/spark-operator/pkg/controller/sparkapplication"

	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
)

func NewSparkOperatorCommand(stopCh <-chan struct{}) *cobra.Command {
	opts := options.NewOptions()

	cmd := &cobra.Command{
		Short: "Launch spark-operator",
		Long:  "Launch spark-operator",
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
	sparkApplicationController, err := sparkapplication.NewController(option)
	if err != nil {
		return err
	}
	klog.Info("starting run server...")
	err = sparkApplicationController.Start(option.ControllerThreads, stopCh)
	if err != nil {
		return err
	}

	<-stopCh

	// INFO: 这里做了queue清理工作，可以借鉴下，不过不是很重要
	klog.Info("Shutting down the Spark Operator")
	sparkApplicationController.Stop()
	return nil
}
