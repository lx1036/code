package cmd

import (
	"k8s-lx1036/k8s/storage/kafka/kubetool/pkg/consumer"
	"os"

	"k8s-lx1036/k8s/storage/kafka/kubetool/pkg/signals"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:    "configmap-secret-controller",
		Short:  "A watcher for your Kubernetes cluster",
		Run:    run,
		PreRun: preRun,
	}

	cmd.PersistentFlags().Bool("debug", false, "Enable debug mode")
	cmd.PersistentFlags().String("configfile", "", "Config File")

	_ = viper.BindPFlag("debug", cmd.Flags().Lookup("debug"))
	_ = viper.BindPFlag("configfile", cmd.Flags().Lookup("configfile"))

	return cmd
}

func preRun(cmd *cobra.Command, args []string) {
	viper.AutomaticEnv()
	configFile := viper.GetString("configfile")
	if len(configFile) != 0 {
		viper.SetConfigFile(configFile)
		if err := viper.ReadInConfig(); err != nil {
			panic(err.Error())
		}
	}

	if viper.GetBool("debug") {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}

	log.SetOutput(os.Stdout)
	log.SetFormatter(&log.JSONFormatter{})
}

func run(cmd *cobra.Command, args []string) {
	stopCh := signals.SetupSignalHandler()

	c, err := consumer.NewConsumer()
	if err != nil {
		log.Errorf("unable to create kafka consumer")
		os.Exit(1)
	}

	if err = c.Run(stopCh); err != nil {
		log.Fatalf("Error running controller: %v", err)
	}

	<-stopCh
	log.Info("shutdown kafka consumer...")
}
