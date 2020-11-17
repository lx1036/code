package controller

import (
	k8s "k8s-lx1036/k8s/storage/log/filebeat/daemonset-operator/controller/kubernetes"
	appsv1 "k8s.io/api/apps/v1"
	coreV1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"time"
)

type LogController struct {
	InformerResources []schema.GroupVersionResource
	ApiServerClient   kubernetes.Interface
	PodStore          cache.Store
	PodInformer       cache.Controller
	NodeStore         cache.Store
	NodeInformer      cache.Controller

	TaskQueue *TaskQueue

	stopCh chan struct{}

	NodeName         string
	ResyncPeriod     time.Duration
	TaskHandlePeriod time.Duration
}

func New(apiServerClient kubernetes.Interface, options Options) *LogController {

	/*ctr := &LogController{
		InformerResources: []schema.GroupVersionResource{
			{
				Group:    coreV1.GroupName,
				Version:  coreV1.SchemeGroupVersion.Version,
				Resource: "pods",
			},
			{
				Group: appsv1.GroupName,
				Version:  coreV1.SchemeGroupVersion.Version,
				Resource: "nodes",
			},
		},
		ApiServerClient:   apiServerClient,
		stopCh:            nil,
	}*/

	ctr := &LogController{
		InformerResources: nil,
		ApiServerClient:   nil,
		stopCh:            nil,
		NodeName:          "",
		ResyncPeriod:      0,
		TaskHandlePeriod:  0,
	}

	ctr.TaskQueue = NewTaskQueue(ctr.syncTask)

	podListWatch := cache.NewListWatchFromClient(ctr.ApiServerClient.CoreV1().RESTClient(), "pods", coreV1.NamespaceAll, fields.OneTermEqualSelector("spec.nodeName", ctr.NodeName))
	ctr.PodStore, ctr.PodInformer = cache.NewInformer(podListWatch, &coreV1.Pod{}, ctr.ResyncPeriod, cache.ResourceEventHandlerFuncs{
		AddFunc:    ctr.AddPod,
		UpdateFunc: ctr.UpdatePod,
		DeleteFunc: ctr.DeletePod,
	})

	nodeListWatch := cache.NewListWatchFromClient(ctr.ApiServerClient.CoreV1().RESTClient(), "nodes", coreV1.NamespaceAll, fields.OneTermEqualSelector("metadata.name", ctr.NodeName))
	ctr.NodeStore, ctr.NodeInformer = cache.NewInformer(nodeListWatch, &coreV1.Node{}, ctr.ResyncPeriod, cache.ResourceEventHandlerFuncs{})

}

func (ctr *LogController) AddPod(obj interface{}) {
	pod := obj.(*coreV1.Pod)
	ctr.TaskQueue.Enqueue(pod)
}
func (ctr *LogController) UpdatePod(oldObj, newObj interface{}) {}
func (ctr *LogController) DeletePod(obj interface{}) {
	pod := obj.(*coreV1.Pod)
	ctr.TaskQueue.Enqueue(pod)
}

// 批量处理pod事件，更新filebeat input.yml
func (ctr *LogController) syncTask(tasks []interface{}) {

}

func (ctr *LogController) Run() {

	go ctr.PodInformer.Run(ctr.stopCh)
	go ctr.NodeInformer.Run(ctr.stopCh)

	go ctr.TaskQueue.Run(ctr.TaskHandlePeriod, ctr.stopCh)

	<-ctr.stopCh
	//k8s.Run(ctr.ApiServerClient, ctr.InformerResources, ctr.stopCh)
}
