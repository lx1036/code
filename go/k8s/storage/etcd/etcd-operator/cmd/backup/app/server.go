package app

import (
	"k8s-lx1036/k8s/storage/etcd/etcd-operator/cmd/backup/app/options"
	controller "k8s-lx1036/k8s/storage/etcd/etcd-operator/pkg/controller/backup"

	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
)

func NewEtcdBackupCommand(stopCh <-chan struct{}) *cobra.Command {
	opts := options.NewOptions()

	cmd := &cobra.Command{
		Short: "Launch etcd-backup-operator",
		Long:  "Launch etcd-backup-operator",
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
	etcdBackupController, err := controller.NewController(option)
	if err != nil {
		return err
	}
	klog.Info("starting run server...")
	err = etcdBackupController.Start(option.ControllerThreads, stopCh)
	if err != nil {
		return err
	}

	<-stopCh

	// INFO: 这里做了queue清理工作，可以借鉴下，不过不是很重要
	klog.Info("Shutting down the etcd backup controller...")
	etcdBackupController.Stop()
	return nil
}
