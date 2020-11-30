package cmd

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s-lx1036/k8s/operator/configmap-secret-controller/k8s-trigger-controller/pkg/controller"
	"k8s-lx1036/k8s/operator/configmap-secret-controller/k8s-trigger-controller/pkg/kube"
	"k8s-lx1036/k8s/operator/configmap-secret-controller/k8s-trigger-controller/pkg/metrics"
	"k8s-lx1036/k8s/operator/configmap-secret-controller/k8s-trigger-controller/pkg/signals"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"os"
	"time"
)

func NewRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:    "configmap-secret-controller",
		Short:  "A watcher for your Kubernetes cluster",
		Run:    startConfigmapSecretControllerCmd,
		PreRun: preRun,
	}

	cmd.PersistentFlags().Bool("debug", false, "Enable debug mode")
	cmd.PersistentFlags().String("namespace", "", "Enable debug mode")
	cmd.PersistentFlags().Duration("sync-period", time.Second*30, "Enable debug mode")

	_ = viper.BindPFlag("debug", cmd.Flags().Lookup("debug"))
	_ = viper.BindPFlag("namespace", cmd.Flags().Lookup("namespace"))
	_ = viper.BindPFlag("sync-period", cmd.Flags().Lookup("sync-period"))

	return cmd
}

func preRun(cmd *cobra.Command, args []string) {
	if viper.GetBool("debug") {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}
	log.SetOutput(os.Stdout)
	log.SetFormatter(&log.JSONFormatter{})
}

func startConfigmapSecretControllerCmd(cmd *cobra.Command, args []string) {
	log.Info("Starting ConfigmapSecret Controller")

	stopCh := signals.SetupSignalHandler()

	// create the clientset
	clientset, err := kube.GetKubernetesClient()
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	var currentNamespace = viper.GetString("namespace")
	if len(currentNamespace) == 0 {
		currentNamespace = metav1.NamespaceAll
		log.Warnf("KUBERNETES_NAMESPACE is unset, will detect changes in all namespaces.")
	}

	collectors := metrics.SetupPrometheusEndpoint()

	informerFactory := informers.NewSharedInformerFactory(clientset, time.Hour)

	c, err := controller.NewController(informerFactory, clientset, collectors, currentNamespace)
	if err != nil {
		log.Errorf("unable to create kubernetes watcher")
		os.Exit(1)
	}

	if err = c.Run(2, stopCh); err != nil {
		log.Fatalf("Error running controller: %s", err.Error())
	}

	/*for resourceType := range kube.ResourceMap {
		c, err := controller.NewController(clientset, resourceType, currentNamespace, collectors)
		if err != nil {
			log.Fatalf("%s", err)
		}

		// Now let's start the controller
		log.Infof("Starting Controller to watch resource type: %s", resourceType)
		go c.Run(1, stop)
	}

	// Wait forever
	<-stop*/
}
