package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s-lx1036/k8s-ui/backend/dashboard/client"
	"k8s-lx1036/k8s-ui/backend/dashboard/router"
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
	fmt.Println("Using config file:", viper.ConfigFileUsed())

	client.DefaultClientManager = client.NewClientManager(viper.GetString("common.kubeconfig"), viper.GetString("common.apiserver-host"))

}

func run(cmd *cobra.Command, args []string) {
	router := router.SetupRouter()
	_ = router.Run(":3456")
}
