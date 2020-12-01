package cmd

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s-lx1036/k8s/operator/configmap-secret-controller/configmap-controller/pkg/controller"
	"k8s-lx1036/k8s/operator/configmap-secret-controller/configmap-controller/pkg/kube"
	"k8s-lx1036/k8s/operator/configmap-secret-controller/configmap-controller/pkg/metrics"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
	"time"
)

func NewRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:    "configmap-controller",
		Short:  "A watcher for your Kubernetes cluster",
		Run:    startConfigmapControllerCmd,
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
	viper.AutomaticEnv()
	if viper.GetBool("debug") {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}
	log.SetOutput(os.Stdout)
	log.SetFormatter(&log.JSONFormatter{})
}

func startConfigmapControllerCmd(cmd *cobra.Command, args []string) {
	log.Info("Starting Configmap Controller")

	// create the clientset
	clientset, err := kube.GetKubernetesClient()
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	collectors := metrics.SetupPrometheusEndpoint()

	stop := make(chan struct{})
	defer close(stop)
	var currentNamespace = viper.GetString("namespace")
	if len(currentNamespace) == 0 {
		currentNamespace = metav1.NamespaceAll
		log.Warnf("KUBERNETES_NAMESPACE is unset, will detect changes in all namespaces.")
	}
	for resourceType := range kube.ResourceMap {
		c, err := controller.NewController(clientset, resourceType, currentNamespace, collectors)
		if err != nil {
			log.Fatalf("%s", err)
		}

		// Now let's start the controller
		log.Infof("Starting Controller to watch resource type: %s", resourceType)
		go c.Run(1, stop)
	}

	// Wait forever
	<-stop
}
