package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"k8s-lx1036/k8s/storage/csi/pluginmanager"
	example_plugin "k8s-lx1036/k8s/storage/csi/pluginmanager/demo/example-plugin"

	"k8s.io/klog/v2"
	registerapi "k8s.io/kubelet/pkg/apis/pluginregistration/v1"
)

var (
	socketDir = flag.String("socketDir", "/tmp/csi_example", "CSI endpoint")

	onlyOneSignalHandler = make(chan struct{})
)

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
		//<-c
		os.Exit(1) // second signal. Exit directly.
	}()

	return stop
}

func newTestPluginManager(sockDir string) pluginmanager.PluginManager {
	return pluginmanager.NewPluginManager(sockDir)
}

// (1)第一步启动pluginmanager，等待plugin来注册，并注册一个csi plugin的handler，用来消费csi plugin
// debug: go run . --socketDir=/tmp/csi_example
func main() {
	klog.InitFlags(nil)
	flag.Set("logtostderr", "true")
	flag.Parse()

	stopCh := SetupSignalHandler()
	pluginMgr := newTestPluginManager(*socketDir)
	go pluginMgr.Run(stopCh)
	//pluginMgr.AddHandler(registerapi.CSIPlugin, csi_plugin.PluginHandler)

	exampleHandler := example_plugin.NewExampleHandler([]string{"v1beta1", "v1beta2"}, true)
	pluginMgr.AddHandler(registerapi.CSIPlugin, exampleHandler)

	<-stopCh
	klog.Info("shutdown the csi plugin manager")
}
