package app

import (
	"fmt"
	"k8s-lx1036/k8s/network/network-policy/cmd/app/options"
	"k8s-lx1036/k8s/network/network-policy/pkg/controller"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"sync"
	"time"

	"github.com/spf13/cobra"
)

func NewNetworkPolicyCommand(stopCh <-chan struct{}) *cobra.Command {
	opts := options.NewOptions()

	cmd := &cobra.Command{
		Short: "Launch network policy contoller",
		Long:  "network policy",
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
	var ipsetMutex sync.Mutex

	restConfig, err := clientcmd.BuildConfigFromFlags("", option.Kubeconfig)
	if err != nil {
		return err
	}
	kubeClient, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return err
	}

	informerFactory := informers.NewSharedInformerFactory(kubeClient, 0)
	networkPolicyInformer := informerFactory.Networking().V1().NetworkPolicies().Informer()
	podInformer := informerFactory.Core().V1().Pods().Informer()
	nsInformer := informerFactory.Core().V1().Namespaces().Informer()
	informerFactory.Start(stopCh)

	// list and watch
	syncOverCh := make(chan struct{})
	go func() {
		informerFactory.WaitForCacheSync(stopCh)
		close(syncOverCh)
	}()
	select {
	case <-time.After(60 * time.Second):
		return fmt.Errorf("timeout")
	case <-syncOverCh:
	}

	c, err := controller.NewNetworkPolicyController(kubeClient, networkPolicyInformer, podInformer, nsInformer, &ipsetMutex)
	if err != nil {
		return err
	}

	go c.Run()

	<-stopCh
	return nil
}
