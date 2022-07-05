package main

import (
	"k8s-lx1036/k8s/network/cilium/cilium/pkg/config/defaults"
	"k8s-lx1036/k8s/network/cilium/cilium/pkg/config/option"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	cobra.OnInitialize(option.InitConfig("ciliumd"))

	flags := RootCmd.Flags()

	////////////////////////////// Base //////////////////////////
	flags.String(option.LibDir, defaults.LibraryPath, "Directory path to store runtime build environment")
	option.BindEnv(option.LibDir)

	////////////////////////////// BPF //////////////////////////
	flags.Bool(option.BPFCompileDebugName, false, "Enable debugging of the BPF compilation process")
	option.BindEnv(option.BPFCompileDebugName)

	flags.Bool(option.InstallIptRules, true, "Install base iptables rules for cilium to mainly interact with kube-proxy (and masquerading)")
	option.BindEnv(option.InstallIptRules)

	viper.BindPFlags(flags)
}

func initEnv() {
	// 从环境变量或cli中，读取配置参数
	option.Config.Populate()

}
