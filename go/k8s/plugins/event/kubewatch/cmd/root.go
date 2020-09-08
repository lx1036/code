package cmd

import (
	"errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s-lx1036/k8s/plugins/event/kubewatch/pkg/client"
	"os"
)

var (
	cfgFile    string
	config     Config
	kubeconfig string
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
	PreRun: preRun,
	Run:    run,
}

func init() {
	RootCmd.Flags().StringVar(&cfgFile, "configfile", "", "config file (default is $HOME/.kubewatch.yaml)")
	RootCmd.Flags().StringVar(&kubeconfig, "kubeconfig", "", "")
	_ = viper.BindPFlag("configfile", RootCmd.Flags().Lookup("configfile"))
	//_ = viper.BindPFlag("kubeconfig", RootCmd.Flags().Lookup("kubeconfig"))
}

func preRun(cmd *cobra.Command, args []string) {
	log.SetOutput(os.Stdout)
	log.SetFormatter(&log.JSONFormatter{})

	//viper.SetConfigName(config.FileName)
	//viper.AddConfigPath("$HOME")
	viper.AutomaticEnv()
	configFile := viper.GetString("configfile")
	if len(configFile) != 0 {
		viper.SetConfigFile(configFile)
		if err := viper.ReadInConfig(); err != nil {
			panic(err.Error())
		}

		if err := viper.Unmarshal(&config); err != nil {
			panic(err.Error())
		}
	} else {
		panic(errors.New("config file is empty"))
	}
}

func run(cmd *cobra.Command, args []string) {
	/*c, err := config.New()
	if err != nil {
		log.WithFields(log.Fields{
			"errmsg": err.Error(),
		}).Error("[rootcmd]")
		os.Exit(1)
	}*/

	/*config := &config.Config{}
	if err := config.Load(); err != nil {
		log.Fatal(err)
	}*/
	//config.CheckMissingResourceEnvvars()
	//c.Run(config)
	//fmt.Println(viper.GetString("namespace"), kubeconfig, config.Namespace)
	//os.Exit(0)
	//controller.Start(config)
	client := client.GetKubeClient(kubeconfig)
	Start(config, client)
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		log.WithFields(log.Fields{
			"errmsg": err.Error(),
		}).Error("[rootcmd]")
		os.Exit(1)
	}
}
