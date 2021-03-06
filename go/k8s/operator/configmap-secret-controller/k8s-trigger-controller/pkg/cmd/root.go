package cmd

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s-lx1036/k8s/operator/configmap-secret-controller/k8s-trigger-controller/pkg/controller"
	"k8s-lx1036/k8s/operator/configmap-secret-controller/k8s-trigger-controller/pkg/health"
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

	cmd.Flags().Bool("debug", false, "enable debug mode")
	cmd.Flags().String("namespace", "", "watch specified namespace")
	cmd.Flags().String("kubeconfig", "", "kubeconfig path file")
	cmd.Flags().Duration("sync-period", time.Second*30, "sync-period for sync resource to local store")

	_ = viper.BindPFlag("debug", cmd.Flags().Lookup("debug"))
	_ = viper.BindPFlag("namespace", cmd.Flags().Lookup("namespace"))
	_ = viper.BindPFlag("kubeconfig", cmd.Flags().Lookup("kubeconfig"))
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

func startConfigmapSecretControllerCmd(cmd *cobra.Command, args []string) {
	log.Info("Starting ConfigmapSecret Controller")

	stopCh := signals.SetupSignalHandler()
	health.SetupHealthCheckHandler()

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

	if err = c.Run(1, stopCh); err != nil {
		log.Fatalf("Error running controller: %s", err.Error())
	}

	<-stopCh
}
