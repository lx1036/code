package cmd

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s-lx1036/k8s/plugins/event/kubewatch/config"
	"os"
)

var RootCmd = &cobra.Command{
	Use:   "k8s_watcher",
	Short: "a watcher for k8s resource",
	Long: `
	k8s_watcher: a watcher for k8s resource.
	supported sink:
		- slack
		- email
		- webhook
`,
	Run: func(cmd *cobra.Command, args []string) {
		c, err := config.New()
		if err != nil {
			log.WithFields(log.Fields{
				"errmsg": err.Error(),
			}).Error("[rootcmd]")
			os.Exit(1)
		}

	},
}

func init() {
	cobra.OnInitialize(initConfig)
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		log.WithFields(log.Fields{
			"errmsg": err.Error(),
		}).Error("[rootcmd]")
		os.Exit(1)
	}
}

func initConfig() {
	log.SetOutput(os.Stdout)
	log.SetFormatter(&log.JSONFormatter{})

	viper.SetConfigName(config.FileName)
	viper.AddConfigPath("$HOME")
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		log.WithFields(log.Fields{
			"errmsg": err.Error(),
		}).Error("[config]")
		os.Exit(1)
	}
}
