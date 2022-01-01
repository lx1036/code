package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"k8s-lx1036/k8s/storage/fusefs/cmd/client"

	"k8s-lx1036/k8s/storage/fuse"
	"k8s-lx1036/k8s/storage/fuse/fuseutil"

	"github.com/spf13/cobra"
	genericapiserver "k8s.io/apiserver/pkg/server"
	"k8s.io/component-base/logs"
	"k8s.io/klog/v2"
)

// INFO: https://chubaofs.readthedocs.io/zh_CN/latest/design/client.html
// go run . --config=./fuse.json
// 检查：df -h globalmount
// 调用 stat 接口：stat globalmount

// debug in local: Working Directory 设置 /Users/liuxiang/Code/lx1036/code/go/k8s/storage/fusefs/cmd/client/cmd
// 关闭时直接执行，进程优雅关闭：`umount globalmount`
func main() {
	logs.InitLogs()
	defer logs.FlushLogs()

	if len(os.Getenv("GOMAXPROCS")) == 0 {
		runtime.GOMAXPROCS(runtime.NumCPU())
	}

	var config string
	cmd := &cobra.Command{
		Use:   "master",
		Short: "Runs the FuseFS client",
		Long:  `responsible for fusefs client`,
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
	var config client.Config
	err = json.Unmarshal(content, &config)
	if err != nil {
		return err
	}

	fuseFS, err := client.NewFuseFS(&config)
	if err != nil {
		fuseFS.Destroy()
		klog.Fatal(err)
	}

	klog.Info("Starting the fuse client")
	mountPoint, _ := filepath.Abs(config.MountPoint)
	mountedFileSystem, err := fuse.Mount(mountPoint, fuseutil.NewFileSystemServer(fuseFS), &fuse.MountConfig{
		FSName:                  "fuse-" + config.Volname,
		Subtype:                 "fuse", // `cat /proc/mounts | grep sunfs` -> xxx fuse.sunfs xxx
		ReadOnly:                config.ReadOnly,
		DisableWritebackCaching: true,
		DebugLogger:             log.New(os.Stderr, "fuse: ", log.LstdFlags),
	})
	if err != nil {
		fuseFS.Destroy()
		klog.Fatal(err)
	}
	if err = mountedFileSystem.Join(genericapiserver.SetupSignalContext()); err != nil {
		klog.Fatal(err)
	}

	return nil
}
