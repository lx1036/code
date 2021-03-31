package app

import (
	"fmt"
	"time"
	
	
	"k8s-lx1036/k8s/monitor/hpa/pkg/podautoscaler"
	"k8s-lx1036/k8s/monitor/hpa/pkg/podautoscaler/metrics"

	"github.com/spf13/cobra"

	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
	cacheddiscovery "k8s.io/client-go/discovery/cached"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/scale"
	"k8s.io/metrics/pkg/client/custom_metrics"
	"k8s.io/metrics/pkg/client/external_metrics"
	resourceclient "k8s.io/metrics/pkg/client/clientset/versioned/typed/metrics/v1beta1"

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
	clientConfig, err := clientcmd.BuildConfigFromFlags("", opts.Kubeconfig)
	if err != nil {
		panic(err)
	}
	clientSet, err := kubernetes.NewForConfig(clientConfig)
	if err != nil {
		panic(err)
	}
	informerFactory := informers.NewSharedInformerFactory(clientSet, time.Minute*5)
	
	scaleKindResolver := scale.NewDiscoveryScaleKindResolver(clientSet.Discovery())
	cachedClient := cacheddiscovery.NewMemCacheClient(clientSet.Discovery())
	restMapper := restmapper.NewDeferredDiscoveryRESTMapper(cachedClient)
	scaleClient, err := scale.NewForConfig(clientConfig, restMapper, dynamic.LegacyAPIPathResolverFunc, scaleKindResolver)
	if err != nil {
		panic(err)
	}
	
	apiVersionsGetter := custom_metrics.NewAvailableAPIsGetter(clientSet.Discovery())
	metricsClient := metrics.NewRESTMetricsClient(
		resourceclient.NewForConfigOrDie(clientConfig),
		custom_metrics.NewForConfig(clientConfig, restMapper, apiVersionsGetter),
		external_metrics.NewForConfigOrDie(clientConfig),
	)
	
	
	
	hpaController := podautoscaler.NewHorizontalController(
		informerFactory.Autoscaling().V1().HorizontalPodAutoscalers(),
		informerFactory.Core().V1().Pods(),
		scaleClient,
		restMapper,
		metricsClient,
	)
	
	
	
	
	stopCh := make(chan struct{})
	
	
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
