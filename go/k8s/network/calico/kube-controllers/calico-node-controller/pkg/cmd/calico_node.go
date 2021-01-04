package cmd

import (
	"os"
	"time"

	"k8s-lx1036/k8s/network/calico/kube-controllers/calico-node-controller/pkg/controller"
	"k8s-lx1036/k8s/network/calico/kube-controllers/calico-node-controller/pkg/signals"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/projectcalico/libcalico-go/lib/apiconfig"
)

func NewNodeControllerCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:    "node-controller",
		Short:  "A watcher for your Kubernetes cluster",
		Run:    startNodeControllerCmd,
		PreRun: preRun,
	}

	cmd.Flags().Bool("debug", false, "enable debug mode")
	cmd.Flags().String("datastore-type", string(apiconfig.Kubernetes), "calico datastore type")
	cmd.Flags().String("namespace", "", "watch specified namespace")
	cmd.Flags().String("kubeconfig", "", "kubeconfig path file")
	cmd.Flags().Duration("sync-period", time.Second*30, "sync-period for sync resource to local store")

	_ = viper.BindPFlag("debug", cmd.Flags().Lookup("debug"))
	_ = viper.BindPFlag("datastore-type", cmd.Flags().Lookup("datastore-type"))
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

func startNodeControllerCmd(cmd *cobra.Command, args []string) {
	log.Info("Starting Node Controller...")

	stopCh := signals.SetupSignalHandler()

	namespaceController := controller.NewNodeController()

	if err := namespaceController.Run(1, stopCh); err != nil {
		log.Fatalf("Error running controller: %s", err.Error())
	}

	<-stopCh

	log.Info("Stopped Node Controller...")
}
