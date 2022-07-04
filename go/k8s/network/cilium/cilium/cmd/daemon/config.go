package main

import (
	"github.com/spf13/viper"
	"k8s-lx1036/k8s/network/cilium/cilium/pkg/config/option"

	"github.com/spf13/cobra"
)

func init() {
	cobra.OnInitialize(option.InitConfig("ciliumd"))

	flags := RootCmd.Flags()

	flags.Bool(option.InstallIptRules, true, "Install base iptables rules for cilium to mainly interact with kube-proxy (and masquerading)")
	option.BindEnv(option.InstallIptRules)

	viper.BindPFlags(flags)
}

func initEnv() {
	// 从环境变量或cli中，读取配置参数
	option.Config.Populate()

}
