package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
	api "k8s.io/kubernetes/pkg/apis/core"
	resourceclient "k8s.io/metrics/pkg/client/clientset/versioned/typed/metrics/v1beta1"
)

// (1) 获取当前 node 的 cpu topo，以及包括 cpu capacity 和 allocatable，参考 kubelet

// (2) 获取当前 node 在线业务的 cpu_usage，参考kubelet

var (
	kubeconfig = flag.String("kubeconfig", "", "absolute path to kubeconfig file")
	nodeName   = flag.String("node", "", "current node")
)

// 启动 metrics-client 来调用 metrics-server 获取 node/pod metrics 数据
// HPA 就是这么做的，使用 pod cpu/memory metrics 来计算replicas副本数量
// debug in local: go run . --kubeconfig=`echo $HOME`/.kube/config --node=docker1234
func main() {
	flag.Parse()

	if len(*kubeconfig) == 0 || len(*nodeName) == 0 {
		klog.Errorf("--kubeconfig or --node should be required")
		return
	}

	clientConfig, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err)
	}

	clientSet, err := kubernetes.NewForConfig(clientConfig)
	if err != nil {
		panic(err)
	}

	// INFO: list 当前 node 上的 pods
	stopCh := context.TODO().Done()
	factory := informers.NewSharedInformerFactoryWithOptions(clientSet, time.Second*10, informers.WithTweakListOptions(func(options *metav1.ListOptions) {
		options.FieldSelector = fields.Set{api.PodHostField: string(*nodeName)}.String()
	}))
	podLister := factory.Core().V1().Pods().Lister()
	factory.Start(stopCh)
	informersSynced := []cache.InformerSynced{
		factory.Core().V1().Pods().Informer().HasSynced,
	}
	if !cache.WaitForCacheSync(stopCh, informersSynced...) {
		klog.Errorf("can not sync pods in node %s", *nodeName)
		return
	}

	pods, err := podLister.Pods(metav1.NamespaceAll).List(labels.Everything())
	if err != nil {
		panic(err)
	}
	klog.Infof("%d pods in node %s", len(pods), *nodeName)
	for _, pod := range pods {
		klog.Infof("%s/%s", pod.Namespace, pod.Name)
	}

	// TODO: 统计该 node 上所有 pod 的 cpu/memory 资源总和
	totalResource := v1.ResourceList{
		v1.ResourceCPU:              *resource.NewMilliQuantity(0, resource.DecimalSI),
		v1.ResourceMemory:           *resource.NewQuantity(0, resource.BinarySI),
		v1.ResourceEphemeralStorage: *resource.NewQuantity(0, resource.BinarySI),
	}
	metricsClient := resourceclient.NewForConfigOrDie(clientConfig)
	// INFO: list pods metrics on current node
	for _, pod := range pods {
		podMetrics, err := metricsClient.PodMetricses(pod.Namespace).Get(context.TODO(), pod.Name, metav1.GetOptions{})
		if err != nil {
			klog.Errorf("fail to get pod %s/%s metrics: %v", pod.Namespace, pod.Name, err)
			continue
		}

		msg := fmt.Sprintf("pod: %s/%s interval:[%s] ", podMetrics.Namespace, podMetrics.Name, podMetrics.Window)
		for _, containerMetrics := range podMetrics.Containers {
			msg += fmt.Sprintf("containerName:%s usage:[cpu %s memory %s] ",
				containerMetrics.Name, containerMetrics.Usage.Cpu(), containerMetrics.Usage.Memory())

			cpu := totalResource[v1.ResourceCPU]
			cpu.Add(*containerMetrics.Usage.Cpu())
			totalResource[v1.ResourceCPU] = cpu
			memory := totalResource[v1.ResourceMemory]
			memory.Add(*containerMetrics.Usage.Memory())
			totalResource[v1.ResourceMemory] = memory
		}

		klog.Info(msg)
	}

	// INFO: get metrics-server pod metrics
	podMetricsList, err := metricsClient.PodMetricses(metav1.NamespaceSystem).List(context.TODO(), metav1.ListOptions{
		LabelSelector: "k8s-app=metrics-server",
	})
	if err != nil {
		panic(err)
	}
	for _, podMetrics := range podMetricsList.Items {
		msg := fmt.Sprintf("podName:%s interval:[%s] ", podMetrics.Name, podMetrics.Window)
		for _, containerMetrics := range podMetrics.Containers {
			msg += fmt.Sprintf("containerName:%s usage:[cpu %s memory %s] ",
				containerMetrics.Name, containerMetrics.Usage.Cpu(), containerMetrics.Usage.Memory())
		}

		klog.Info(msg)
	}

	// INFO: total resource in node
	totalCpu := totalResource[v1.ResourceCPU]
	totalMemory := totalResource[v1.ResourceMemory]
	klog.Infof("total resource cpu: %s, memory: %s in node %s", totalCpu.String(), totalMemory.String(), *nodeName)
}
