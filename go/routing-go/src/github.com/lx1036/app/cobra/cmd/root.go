package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
)

var configFile string

var rootCmd = &cobra.Command{
	Use: "hello",
	Short: "for test",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Hello cobra")
	},
}

var versionCmd = &cobra.Command{
	Use: "version",
	Short: "Print the version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("version: 1.0.0")
	},
}

func Execute()  {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init()  {
	cobra.OnInitialize(initConfig)

	rootCmd.AddCommand(versionCmd)
}

func initConfig()  {
	if configFile != "" {
		viper.SetConfigFile(configFile)
	} else {

	}
}
