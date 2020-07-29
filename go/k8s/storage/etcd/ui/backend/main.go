package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s-lx1036/k8s/storage/etcd/ui/backend/router"
	"os"
)

var (
	// ProjectName : 项目名称
	ProjectName string
	// Version : 版本信息
	Version string
	// Debug : 是否开启Debug
	Debug bool
	// Port : 服务启动的端口号
	Port int
)

func main() {
	Version = "1.0.0"
	ProjectName = os.Getenv("PROJECT_NAME")
	var rootCmd = &cobra.Command{
		Use:    ProjectName,
		Short:  fmt.Sprintf("%s %s", ProjectName, Version),
		Run:    run,
		PreRun: preRun,
	}
	rootCmd.Flags().Bool("debug", false, "Enable debug mode")
	rootCmd.Flags().Int("port", 8080, "Listen port")
	rootCmd.Flags().String("configfile", "", "Config File")

	_ = viper.BindPFlag("debug", rootCmd.Flags().Lookup("debug"))
	_ = viper.BindPFlag("port", rootCmd.Flags().Lookup("port"))
	_ = viper.BindPFlag("configfile", rootCmd.Flags().Lookup("configfile"))

	_ = rootCmd.Execute()
}

func preRun(cmd *cobra.Command, args []string) {
	loadEnvironment()

	Debug = viper.GetBool("debug")
	Port = viper.GetInt("port")

	if Debug {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}

	log.SetOutput(os.Stdout)
	log.SetFormatter(&log.JSONFormatter{})

}

func loadEnvironment() {
	viper.AutomaticEnv()
	configFile := viper.GetString("configfile")
	if len(configFile) != 0 {
		viper.SetConfigFile(configFile)
		if err := viper.ReadInConfig(); err != nil {
			panic(err.Error())
		}
	}
}

func run(cmd *cobra.Command, args []string) {
	app := router.SetupRouter()
	_ = app.Run(fmt.Sprintf(":%d", Port))
}
