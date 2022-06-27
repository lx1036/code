package app

import "k8s-lx1036/k8s/network/cilium/cilium/daemon/pkg/k8s/watchers"

type Daemon struct {
	k8sWatcher *watchers.K8sWatcher
}

func NewDaemon() (*Daemon, error) {

	d := Daemon{}

	d.k8sWatcher = watchers.NewK8sWatcher()

	d.k8sCachesSynced = d.k8sWatcher.InitK8sSubsystem()

}
