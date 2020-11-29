package controller

import (
	"context"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"k8s-lx1036/k8s/storage/log/filebeat/daemonset-operator/common"
	k8s "k8s-lx1036/k8s/storage/log/filebeat/daemonset-operator/controller/kubernetes"
	coreV1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"os"
	"strings"
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

const defaultNode = "localhost"

func InClusterNamespace() (string, error) {
	data, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

func DiscoverKubernetesNode(host string, client kubernetes.Interface) string {
	if len(host) != 0 {
		log.Infof("Using node %s provided by env", host)
		return host
	}

	// node discover by pod
	ns, err := InClusterNamespace()
	if err != nil {
		log.Errorf("Can't get namespace in cluster with error: %v", err)
		return defaultNode
	}
	podName, err := os.Hostname()
	if err != nil {
		log.Errorf("Can't get hostname as pod name in cluster with error: %v", err)
		return defaultNode
	}

	pod, err := client.CoreV1().Pods(ns).Get(context.TODO(), podName, metav1.GetOptions{})
	if err != nil {
		log.Errorf("Can't query pod in cluster with error: %v", err)
		return defaultNode
	}

	log.Infof("Using node %s discovered by pod in cluster", pod.Spec.NodeName)

	return pod.Spec.NodeName
}

func New(options common.Options) *LogController {

	client, err := common.GetKubernetesClient(options.KubeConfig)
	if err != nil {

	}

	podWatcher, err := k8s.NewWatcher(client, &k8s.Pod{}, k8s.WatchOptions{
		SyncTimeout: time.Minute * 10,
		Node:        DiscoverKubernetesNode(options.Host, client),
		Namespace:   "",
		IsUpdated:   nil,
	}, nil)

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
