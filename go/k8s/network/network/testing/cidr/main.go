package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"

	"k8s.io/klog/v2"
)

// go run . --cluster-pool-ipv4-cidr="1.2.3.4/24,2.2.3.4/24"
func main() {
	const ClusterPoolIPv4CIDR = "cluster-pool-ipv4-cidr"
	rootCmd := &cobra.Command{
		Short: "Run ",
		Run: func(cmd *cobra.Command, args []string) {
			klog.Info("success")

			cidr := viper.GetStringSlice(ClusterPoolIPv4CIDR)
			for _, value := range cidr {
				klog.Info(fmt.Sprintf("cidr: %s", value))
				//I0831 18:14:20.944470   35862 main.go:18] success
				//I0831 18:14:20.944584   35862 main.go:22] cidr: 1.2.3.4/24
				//I0831 18:14:20.944591   35862 main.go:22] cidr: 2.2.3.4/24
			}
		},
	}
	flags := rootCmd.Flags()
	flags.StringSlice(ClusterPoolIPv4CIDR, []string{}, "")
	//pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	//pflag.Parse()
	viper.BindPFlags(flags)
	//viper.BindPFlags(pflag.CommandLine)
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}
