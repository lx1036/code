package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s-lx1036/k8s-ui/dashboard/client"
	"k8s-lx1036/k8s-ui/dashboard/router"
	"os"
	"path/filepath"
)

const (
	Version     = "1.6.1"
	ProjectName = "k8s-ui"
)

var (
	configFile string
)

func main() {
	log.SetOutput(os.Stdout)
	log.SetFormatter(&log.JSONFormatter{})

	var rootCmd = &cobra.Command{
		Use:    ProjectName,
		Short:  fmt.Sprintf("%s %s", ProjectName, Version),
		Run:    run,
		PreRun: preRun,
	}

	rootCmd.PersistentFlags().StringVar(&configFile, "configFile", "app.conf", "config file path")

	_ = rootCmd.Execute()
}

func preRun(cmd *cobra.Command, args []string) {
	filename, _ := filepath.Abs(".")
	viper.SetConfigType("ini")
	file := fmt.Sprintf("%s/conf/%s", filename, configFile)
	viper.SetConfigFile(file)
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}
	
	log.WithFields(log.Fields{
		"config-file": viper.ConfigFileUsed(),
	}).Info("[app level]")

	client.DefaultClientManager = client.NewClientManager(viper.GetString("common.kubeconfig"), viper.GetString("common.apiserver-host"))
}

func run(cmd *cobra.Command, args []string) {
	app := router.SetupRouter()
	err := app.Run(":3456")
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Info("[app level]")
	}
}
