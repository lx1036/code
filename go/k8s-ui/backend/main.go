package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	_ "k8s-lx1036/k8s-ui/backend/database/lorm"
	"k8s-lx1036/k8s-ui/backend/initial"
	routers_gin "k8s-lx1036/k8s-ui/backend/routers-gin"
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

	/*cmd.Version = Version
	_ = cmd.RootCmd.Execute()*/

	//cmd2.Run()

	//database.InitDb()

	// K8S Client
	//initial.InitClient()

	// 初始化 rsa key
	_ = rootCmd.Execute()

}

func run(cmd *cobra.Command, args []string) {
	router := routers_gin.SetupRouter()
	_ = router.Run(":8080")
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
}
