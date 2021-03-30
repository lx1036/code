package podautoscaler

import (
	"time"

	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	autoscalinginformers "k8s.io/client-go/informers/autoscaling/v1"
	coreinformers "k8s.io/client-go/informers/core/v1"
	autoscalinglisters "k8s.io/client-go/listers/autoscaling/v1"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
)

type HorizontalController struct {

	// autoscalinglisters 这里使用的是v1版本，代码里会转换v2beta2
	hpaLister       autoscalinglisters.HorizontalPodAutoscalerLister
	hpaListerSynced cache.InformerSynced
	podLister       corelisters.PodLister
	podListerSynced cache.InformerSynced
}

// NewHorizontalController creates a new HorizontalController.
func NewHorizontalController(
	hpaInformer autoscalinginformers.HorizontalPodAutoscalerInformer,
	podInformer coreinformers.PodInformer,
) *HorizontalController {

	hpaController := &HorizontalController{}

	hpaInformer.Informer().AddEventHandlerWithResyncPeriod(cache.ResourceEventHandlerFuncs{
		AddFunc:    nil,
		UpdateFunc: nil,
		DeleteFunc: nil,
	}, time.Second*10)
	hpaController.hpaLister = hpaInformer.Lister()
	hpaController.hpaListerSynced = hpaInformer.Informer().HasSynced
	hpaController.podLister = podInformer.Lister()
	hpaController.podListerSynced = podInformer.Informer().HasSynced

	return hpaController
}

func (hpa *HorizontalController) Run(stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()
	defer hpa.queue.ShutDown()

	klog.Infof("Starting HPA controller")
	defer klog.Infof("Shutting down HPA controller")

	if !cache.WaitForNamedCacheSync("HPA", stopCh, a.hpaListerSynced, a.podListerSynced) {
		return
	}

	// start a single worker (we may wish to start more in the future)
	go wait.Until(hpa.worker, time.Second, stopCh)

	<-stopCh
}
