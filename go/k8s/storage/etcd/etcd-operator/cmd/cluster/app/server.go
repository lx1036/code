package app

import (
	"k8s-lx1036/k8s/storage/etcd/etcd-operator/cmd/cluster/app/options"
	"k8s-lx1036/k8s/storage/etcd/etcd-operator/pkg/controller"

	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
)

func NewEtcdOperatorCommand(stopCh <-chan struct{}) *cobra.Command {
	opts := options.NewOptions()

	cmd := &cobra.Command{
		Short: "Launch etcd-cluster-operator",
		Long:  "Launch etcd-cluster-operator",
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
	etcdClusterController, err := controller.NewController(option)
	if err != nil {
		return err
	}
	klog.Info("starting run server...")
	err = etcdClusterController.Start(option.ControllerThreads, stopCh)
	if err != nil {
		return err
	}

	<-stopCh

	// INFO: 这里做了queue清理工作，可以借鉴下，不过不是很重要
	klog.Info("Shutting down the etcd cluster")
	etcdClusterController.Stop()
	return nil
}
