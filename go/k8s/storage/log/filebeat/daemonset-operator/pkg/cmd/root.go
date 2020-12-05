package cmd

import (
	"os"
	"time"

	"k8s-lx1036/k8s/storage/log/filebeat/daemonset-operator/pkg/controller"
	"k8s-lx1036/k8s/storage/log/filebeat/daemonset-operator/pkg/health"
	"k8s-lx1036/k8s/storage/log/filebeat/daemonset-operator/pkg/kube"
	"k8s-lx1036/k8s/storage/log/filebeat/daemonset-operator/pkg/metrics"
	"k8s-lx1036/k8s/storage/log/filebeat/daemonset-operator/pkg/signals"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
)

func NewRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:    "filebeat-controller",
		Short:  "A watcher for your Kubernetes cluster",
		Run:    startFilebeatControllerCmd,
		PreRun: preRun,
	}

	cmd.Flags().Bool("debug", false, "enable debug mode")
	cmd.Flags().String("node", "", "specified host node")
	cmd.Flags().String("namespace", "", "watch specified namespace")
	cmd.Flags().String("kubeconfig", "", "kubeconfig path file")
	cmd.Flags().Duration("sync-period", 10*time.Minute, "sync-period for sync resource to local store")

	_ = viper.BindPFlag("debug", cmd.Flags().Lookup("debug"))
	_ = viper.BindPFlag("node", cmd.Flags().Lookup("node"))
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

func startFilebeatControllerCmd(cmd *cobra.Command, args []string) {
	log.Info("Starting Filebeat Controller")

	stopCh := signals.SetupSignalHandler()
	health.SetupHealthCheckHandler()

	// create the clientset
	clientset, err := kube.GetKubernetesClient()
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	if len(viper.GetString("namespace")) == 0 {
		//currentNamespace = metav1.NamespaceAll
		viper.Set("namespace", metav1.NamespaceAll)
		log.Warnf("KUBERNETES_NAMESPACE is unset, will detect changes in all namespaces.")
	}

	collectors := metrics.SetupPrometheusEndpoint()
	informerFactory := informers.NewSharedInformerFactory(clientset, time.Hour)
	c, err := controller.NewController(informerFactory, clientset, collectors)
	if err != nil {
		log.Errorf("unable to create kubernetes watcher")
		os.Exit(1)
	}

	if err = c.Run(1, stopCh); err != nil {
		log.Fatalf("Error running controller: %s", err.Error())
	}

	<-stopCh
}
