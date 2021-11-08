package app

import (
	"fmt"
	"k8s-lx1036/k8s/network/kube-router/cmd/app/options"
	"k8s-lx1036/k8s/network/kube-router/pkg/controllers/routing"
	"k8s-lx1036/k8s/network/kube-router/pkg/utils"
	"k8s.io/client-go/kubernetes"
	"time"

	"github.com/spf13/cobra"
	"k8s.io/client-go/informers"
	"k8s.io/klog/v2"
)

func NewKubeRouterCommand(stopCh <-chan struct{}) *cobra.Command {
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
	clientConfig, err := utils.NewRestConfig(option.Kubeconfig)
	if err != nil {
		return err
	}
	clientSet, err := kubernetes.NewForConfig(clientConfig)
	if err != nil {
		return err
	}

	factory := informers.NewSharedInformerFactory(clientSet, 0)
	nodeInformer := factory.Core().V1().Nodes().Informer()
	serviceInformer := factory.Core().V1().Services().Informer()
	endpointInformer := factory.Core().V1().Endpoints().Informer()
	factory.Start(stopCh)
	syncCh := make(chan struct{})
	go func() {
		factory.WaitForCacheSync(stopCh)
		close(syncCh)
	}()
	t := time.Second * 60
	select {
	case <-time.After(t):
		return fmt.Errorf(fmt.Sprintf("timeout %s for sync cache", t.String()))
	case <-syncCh:
	}

	controller, err := routing.NewNetworkRoutingController(option)
	if err != nil {
		return err
	}
	klog.Info("starting run server...")
	err = controller.Start(option.ControllerThreads, stopCh)
	if err != nil {
		return err
	}

	<-stopCh

	// INFO: 这里做了queue清理工作，可以借鉴下，不过不是很重要
	klog.Info("Shutting down the etcd cluster")
	controller.Stop()
	return nil
}
