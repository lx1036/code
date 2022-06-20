package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	genericapiserver "k8s.io/apiserver/pkg/server"
	"os"
	"path/filepath"
	"runtime"

	"k8s-lx1036/k8s/storage/fusefs/cmd/server/master"

	"github.com/spf13/cobra"
	"k8s.io/component-base/logs"
	"k8s.io/klog/v2"
)

// TODO: 把 tiglabs/raft 替换为 hashicorp/raft

// go run . --config=`pwd`/master.json
func main() {
	logs.InitLogs()
	defer logs.FlushLogs()

	if len(os.Getenv("GOMAXPROCS")) == 0 {
		runtime.GOMAXPROCS(runtime.NumCPU())
	}

	var config string

	cmd := &cobra.Command{
		Use:   "master",
		Short: "Runs the FuseFS master server",
		Long:  `responsible for volume creation, query and deletion, node heartbeat state detection, etc`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(config) == 0 {
				klog.Fatal("config is required")
			}
			if err := runCommand(config); err != nil {
				return err
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&config, "config", "", "master config file")
	cmd.Flags().AddGoFlagSet(flag.CommandLine)

	if err := cmd.Execute(); err != nil {
		klog.Fatal(err)
	}
}

func runCommand(configFile string) error {
	configFile, _ = filepath.Abs(configFile)
	content, err := ioutil.ReadFile(configFile)
	if err != nil {
		return err
	}
	var config master.Config
	err = json.Unmarshal(content, &config)
	if err != nil {
		return err
	}

	stopCh := genericapiserver.SetupSignalHandler()
	server := master.NewServer(config)
	err = server.Start()
	if err != nil {
		return err
	}

	klog.Info("Starting the master raft cluster")

	<-stopCh

	klog.Info("Shutting down the master raft cluster")
	server.Stop()
	return nil
}
