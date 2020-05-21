package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s-lx1036/k8s-ui/backend/client"
	_ "k8s-lx1036/k8s-ui/backend/database/lorm"
	"k8s-lx1036/k8s-ui/backend/initial"
	routers_gin "k8s-lx1036/k8s-ui/backend/routers-gin"
	"k8s.io/apimachinery/pkg/util/wait"
	"os"
	"path/filepath"
	"time"
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

	// 初始化 rsa key
	err := rootCmd.Execute()
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Info("[app level]")
	}
}

func run(cmd *cobra.Command, args []string) {
	router := routers_gin.SetupRouter()
	_ = router.Run(":3456")
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

	initial.InitRsaKey(viper.GetString("default.RsaPrivateKey"), viper.GetString("default.RsaPublicKey"))

	// K8S Client
	go wait.Forever(client.BuildApiServerClient, 30*time.Second)
}
