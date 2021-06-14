package app

import (
	"k8s-lx1036/k8s/monitor/vpa/recommender/cmd/app/options"
	"k8s-lx1036/k8s/monitor/vpa/recommender/pkg/controller/recommender"

	"github.com/spf13/cobra"
)

func NewRecommenderCommand(stopCh <-chan struct{}) *cobra.Command {
	opts := options.NewOptions()

	cmd := &cobra.Command{
		Short: "Launch recommender",
		Long:  "Launch recommender",
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
	r, err := recommender.NewRecommender(option)
	if err != nil {
		return err
	}

	return r.RunUntil(stopCh)
}
