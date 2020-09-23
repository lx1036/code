package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s-lx1036/k8s-ui/backend/client"
	"k8s-lx1036/k8s-ui/backend/common/rsa"
	"k8s-lx1036/k8s-ui/backend/common/util"
	"k8s-lx1036/k8s-ui/backend/database"
	routersGin "k8s-lx1036/k8s-ui/backend/routers-gin"
	"k8s.io/apimachinery/pkg/util/wait"
	"os"
	"time"
)

const (
	Version     = "1.6.1"
	ProjectName = "k8s-ui"
)

var (
	configFile string
)

// go run . --configfile=app.conf --port=8080
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
	rootCmd.Flags().Int("port", 8080, "Listen port")

	_ = viper.BindPFlag("configfile", rootCmd.Flags().Lookup("configfile"))
	_ = viper.BindPFlag("port", rootCmd.Flags().Lookup("port"))

	// 初始化 rsa key
	err := rootCmd.Execute()
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Info("[app level]")
	}
}

func run(cmd *cobra.Command, args []string) {
	router := routersGin.SetupRouter()
	db := database.InitDb()
	defer db.Close()

	_ = router.Run(fmt.Sprintf(":%d", viper.GetInt("port")))
}

func preRun(cmd *cobra.Command, args []string) {
	viper.AutomaticEnv()
	configFile := viper.GetString("configfile")
	if len(configFile) != 0 {
		viper.SetConfigFile(configFile)
		if err := viper.ReadInConfig(); err != nil {
			panic(err.Error())
		}
	} else {
		panic(fmt.Sprint("config file can't be empty"))
	}
	log.WithFields(log.Fields{
		"configfile": viper.ConfigFileUsed(),
	}).Info("[configfile]")

	rsa.InitRsaKey()
	if util.AppLabelKey = viper.GetString("default.AppLabelKey"); len(util.AppLabelKey) == 0 {
		util.AppLabelKey = "k8s-app"
	}
	if util.NamespaceLabelKey = viper.GetString("default.NamespaceLabelKey"); len(util.NamespaceLabelKey) == 0 {
		util.NamespaceLabelKey = "k8s-ns"
	}
	if util.PodAnnotationControllerKindLabelKey = viper.GetString("default.PodAnnotationControllerKindLabelKey"); len(util.PodAnnotationControllerKindLabelKey) == 0 {
		util.PodAnnotationControllerKindLabelKey = "k8s.cloud/controller-kind"
	}

	// K8S Client
	go wait.Forever(client.BuildApiServerClient, 30*time.Second)
}
