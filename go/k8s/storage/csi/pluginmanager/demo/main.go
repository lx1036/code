package main

import (
	"k8s-lx1036/k8s/storage/csi/pluginmanager"
	csi_handler "k8s-lx1036/k8s/storage/csi/pluginmanager/demo/csi-handler"
	"k8s.io/klog"
	"os"
	"os/signal"
	"syscall"

	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/record"
	"k8s.io/kubernetes/pkg/kubelet/config"
)

func SeenAllSources(seenSources sets.String) bool {
	return true
}

var onlyOneSignalHandler = make(chan struct{})

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

func main() {
	stopCh := SetupSignalHandler()
	sockDir := "/tmp/csi"
	pluginMgr := pluginmanager.NewPluginManager(sockDir, record.NewFakeRecorder(100))
	pluginMgr.AddHandler("csi", csi_handler.PluginHandler)

	go pluginMgr.Run(config.NewSourcesReady(SeenAllSources), wait.NeverStop)

	<-stopCh
	klog.Info("shutdown the csi plugin manager")
}
