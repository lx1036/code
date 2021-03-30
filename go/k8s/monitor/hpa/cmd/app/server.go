package app

import (
	"fmt"
	"time"

	"k8s-lx1036/k8s/monitor/hpa/pkg/podautoscaler"

	"github.com/spf13/cobra"

	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
)

func NewHPACommand() *cobra.Command {
	opts, err := NewOptions()
	if err != nil {
		klog.Fatalf("unable to initialize command options: %v", err)
	}

	cmd := &cobra.Command{
		Use: "hpa",
		Args: func(cmd *cobra.Command, args []string) error {
			for _, arg := range args {
				if len(arg) > 0 {
					return fmt.Errorf("%q does not take any arguments, got %q", cmd.CommandPath(), args)
				}
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			klog.Fatalf("opts: %v", *opts)

			if err := runCommand(cmd, opts); err != nil {
				klog.Fatal(err)
			}
		},
	}

	fs := cmd.Flags()
	namedFlagSets := opts.Flags()
	for _, f := range namedFlagSets.FlagSets {
		fs.AddFlagSet(f)
	}

	return cmd
}

func runCommand(cmd *cobra.Command, opts *Options) error {
	clientSet := getKubeClient(opts.Kubeconfig)
	informerFactory := informers.NewSharedInformerFactory(clientSet, time.Minute*5)

	stopCh := make(chan struct{})

	hpaController := podautoscaler.NewHorizontalController(
		informerFactory.Autoscaling().V1().HorizontalPodAutoscalers(),
		informerFactory.Core().V1().Pods(),
	)

	informerFactory.Start(stopCh)

	hpaController.Run(stopCh)

}

func getKubeClient(kubeconfig string) *kubernetes.Clientset {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		panic(err)
	}
	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	return clientSet
}
