package app

import (
	"fmt"
	"k8s-lx1036/k8s/network/loadbalancer/kube-router/cmd/app/options"
	"k8s-lx1036/k8s/network/loadbalancer/kube-router/pkg/controllers/routing"
	"k8s-lx1036/k8s/network/loadbalancer/kube-router/pkg/utils"
	"k8s.io/client-go/kubernetes"

	"github.com/spf13/cobra"
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

	controller, err := routing.NewNetworkRoutingController(option)
	if err != nil {
		return err
	}
	klog.Info("starting run server...")
	go controller.Run(stopCh) // INFO: 注意这里是异步

	controller.CondMutex.L.Lock()
	controller.CondMutex.Wait() // INFO: sync.Cond 等待 Start() 里 Cond.Broadcast()，这个编码技巧可以复制!!!
	klog.Infof(fmt.Sprintf("wait for the pod networking related firewall rules to be setup before network policies"))
	controller.CondMutex.L.Unlock()

	<-stopCh

	// INFO: 这里做了queue清理工作，可以借鉴下，不过不是很重要
	klog.Info("Shutting down the etcd cluster")
	controller.Stop()
	return nil
}
