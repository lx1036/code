package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"k8s.io/klog/v2"
	registerapi "k8s.io/kubelet/pkg/apis/pluginregistration/v1"
)

var (
	socketDir = flag.String("socketDir", "/tmp/csi_example", "CSI endpoint")

	onlyOneSignalHandler = make(chan struct{})
)

// INFO: reconcile还有个bug，使用os.RemoveAll(socketPath)删除socket时候，因为fsnotify没有上报删除事件，plugin-watcher没有从desiredStateOfWorld中删除plugin
func cleanup(socketPath string) {
	klog.Infof("cleanup socketDir %s ...", socketPath)
	os.RemoveAll(socketPath)
	//os.RemoveAll(*socketDir)
	//os.MkdirAll(*socketDir, 0755)
}

// SetupSignalHandler registered for SIGTERM and SIGINT. A stop channel is returned
// which is closed on one of these signals. If a second signal is caught, the program
// is terminated with exit code 1.
func SetupSignalHandler() (stopCh <-chan struct{}) {
	close(onlyOneSignalHandler) // panics when called twice

	stop := make(chan struct{})
	c := make(chan os.Signal, 2)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-c
		close(stop)
		<-c
		os.Exit(1) // second signal. Exit directly.
	}()

	return stop
}

// (2)第二步注册一个csi plugin，该plugin会被pluginmanager框架内的csi plugin handler消费
// debug: go run . --socketDir=/tmp/csi_example
func main() {
	//defer cleanup()

	klog.InitFlags(nil)
	flag.Set("logtostderr", "true")
	flag.Parse()

	stopCh := SetupSignalHandler()

	// 启动gRPC服务端
	supportedVersions := []string{"v1beta1", "v1beta2"}
	socketPath := fmt.Sprintf("%s/plugin.sock", *socketDir)
	defer cleanup(socketPath)
	pluginName := "example-plugin"
	plugin := NewTestExamplePlugin(pluginName, registerapi.CSIPlugin, socketPath, supportedVersions...)
	if err := plugin.Serve("v1beta1", "v1beta2"); err != nil {
		klog.Error(err)
		return
	}

	<-stopCh
	//cleanup(socketPath)
	klog.Info("shutdown the csi plugin server")
}
