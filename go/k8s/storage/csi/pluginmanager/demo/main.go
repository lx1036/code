package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
	
	"k8s-lx1036/k8s/storage/csi/pluginmanager"
	example_plugin "k8s-lx1036/k8s/storage/csi/pluginmanager/demo/example-plugin"
	
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog"
	registerapi "k8s.io/kubelet/pkg/apis/pluginregistration/v1"
	"k8s.io/kubernetes/pkg/kubelet/config"
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
	pm := pluginmanager.NewPluginManager(
		sockDir,
		&record.FakeRecorder{},
	)
	return pm
}

func cleanup() {
	os.RemoveAll(*socketDir)
	os.MkdirAll(*socketDir, 0755)
}

func startGrpcServer()  {
	// 启动gRPC服务端
	supportedVersions := []string{"v1beta1", "v1beta2"}
	socketPath := fmt.Sprintf("%s/plugin.sock", *socketDir)
	pluginName := "example-plugin"
	plugin := example_plugin.NewTestExamplePlugin(pluginName, registerapi.CSIPlugin, socketPath, supportedVersions...)
	if err := plugin.Serve("v1beta1", "v1beta2"); err != nil {
		klog.Error(err)
		return
	}
}

// debug: go run . --socketDir=/tmp/csi_example
func main() {
	defer cleanup()
	
	klog.InitFlags(nil)
	flag.Set("logtostderr", "true")
	flag.Parse()
	
	go startGrpcServer()
	time.Sleep(time.Second * 3)

	stopCh := SetupSignalHandler()
	pluginMgr := newTestPluginManager(*socketDir)
	go func() {
		sourcesReady := config.NewSourcesReady(func(_ sets.String) bool { return true })
		pluginMgr.Run(sourcesReady, stopCh)
	}()
	//pluginMgr.AddHandler(registerapi.CSIPlugin, csi_plugin.PluginHandler)

	exampleHandler := example_plugin.NewExampleHandler([]string{"v1beta1", "v1beta2"}, true)
	pluginMgr.AddHandler(registerapi.CSIPlugin, exampleHandler)

	<-stopCh
	klog.Info("shutdown the csi plugin manager")
}
