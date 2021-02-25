package controller

import (
	"context"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
)

func NewResizeController(
	name string,
	resizer resizer.Resizer,
	kubeClient kubernetes.Interface,
	resyncPeriod time.Duration,
	informerFactory informers.SharedInformerFactory,
	pvcRateLimiter workqueue.RateLimiter,
	handleVolumeInUseError bool) ResizeController {

}

// Run starts the controller.
func (ctrl *resizeController) Run(workers int, ctx context.Context) {
	defer ctrl.claimQueue.ShutDown()

	klog.Infof("Starting external resizer %s", ctrl.name)
	defer klog.Infof("Shutting down external resizer %s", ctrl.name)

	stopCh := ctx.Done()
	informersSyncd := []cache.InformerSynced{ctrl.pvSynced, ctrl.pvcSynced}
	if ctrl.handleVolumeInUseError {
		informersSyncd = append(informersSyncd, ctrl.podListerSynced)
	}

	if !cache.WaitForCacheSync(stopCh, informersSyncd...) {
		klog.Errorf("Cannot sync pod, pv or pvc caches")
		return
	}

	for i := 0; i < workers; i++ {
		go wait.Until(ctrl.syncPVCs, 0, stopCh)
	}

	<-stopCh
}
