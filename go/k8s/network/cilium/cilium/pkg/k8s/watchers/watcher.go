package watchers

import (
	"k8s.io/klog/v2"
	"time"
)

const (
	cacheSyncTimeout = 3 * time.Minute
)

type K8sWatcher struct {
}

func NewK8sWatcher() *K8sWatcher {

}

func (k *K8sWatcher) InitK8sSubsystem() <-chan struct{} {
	if err := k.EnableK8sWatcher(); err != nil {
		klog.Fatal("Unable to establish connection to Kubernetes apiserver")
	}

	cachesSynced := make(chan struct{})

	go func() {
		// wait for cache sync data from api-server
		klog.Info("Waiting until all pre-existing resources related to policy have been received")

		close(cachesSynced)
	}()

	go func() {
		select {
		case <-cachesSynced:
			klog.Info("All pre-existing resources related to policy have been received; continuing")
		case <-time.After(cacheSyncTimeout):
			klog.Fatalf("Timed out waiting for pre-existing resources related to policy to be received; exiting")
		}
	}()

	return cachesSynced
}

// EnableK8sWatcher watch k8s service/endpoint/networkpolicy
func (k *K8sWatcher) EnableK8sWatcher() error {

}
