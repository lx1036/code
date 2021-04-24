package main

import (
	"context"
	"flag"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
	resourceclient "k8s.io/metrics/pkg/client/clientset/versioned/typed/metrics/v1beta1"
)

// (1) 获取当前 node 的 cpu topo，以及包括 cpu capacity 和 allocatable，参考 kubelet

// (2) 获取当前 node 在线业务的 cpu_usage，参考kubelet

var (
	kubeconfig = flag.String("kubeconfig", "", "absolute path to kubeconfig file")
)

// 启动 metrics-client 来调用 metrics-server 获取 node/pod metrics 数据
// HPA 就是这么做的，使用 pod cpu/memory metrics 来计算replicas副本数量
// debug in local: go run . --kubeconfig=`echo $HOME`/.kube/config
func main() {
	flag.Parse()

	clientConfig, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err)
	}

	/*clientSet, err := kubernetes.NewForConfig(clientConfig)
	if err != nil {
		panic(err)
	}*/

	// TODO: list pods on current node
	metricsClient := resourceclient.NewForConfigOrDie(clientConfig)
	podMetricsList, err := metricsClient.PodMetricses("kube-system").List(context.TODO(), metav1.ListOptions{
		LabelSelector: "k8s-app=metrics-server",
	})
	if err != nil {
		panic(err)
	}
	for _, podMetrics := range podMetricsList.Items {
		msg := fmt.Sprintf("podName:%s interval:[%s %s] ", podMetrics.Name, podMetrics.Timestamp, podMetrics.Window)
		for _, containerMetrics := range podMetrics.Containers {
			msg += fmt.Sprintf("containerName:%s usage:[cpu %s memory %s] ",
				containerMetrics.Name, containerMetrics.Usage.Cpu(), containerMetrics.Usage.Memory())
		}
		klog.Info(msg)
	}

}
