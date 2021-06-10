package app

import (
	"k8s-lx1036/k8s/bigdata/spark-on-k8s/spark-operator/cmd/app/options"

	"github.com/spf13/cobra"
)

func NewSparkOperatorCommand(stopCh <-chan struct{}) *cobra.Command {
	opts := options.NewOptions()

	cmd := &cobra.Command{
		Short: "Launch metrics-server",
		Long:  "Launch metrics-server",
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

}
