package main

import (
	"context"
	"flag"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
	"os"

	"k8s.io/client-go/tools/clientcmd"
	resourceclient "k8s.io/metrics/pkg/client/clientset/versioned/typed/metrics/v1beta1"
)

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

	clientSet, err := kubernetes.NewForConfig(clientConfig)
	if err != nil {
		panic(err)
	}

	hpa, err := clientSet.AutoscalingV2beta2().HorizontalPodAutoscalers(metav1.NamespaceDefault).Get(context.TODO(), "hpa-test-memory", metav1.GetOptions{})
	if err != nil {
		panic(err)
	}
	if hpa != nil {
		klog.Infof("hpa behavior", hpa.Spec.Behavior.String())
		os.Exit(0)
	}

	resourceMetricsClient := resourceclient.NewForConfigOrDie(clientConfig)

	// 等于 http://127.0.0.1:8001/apis/metrics.k8s.io/v1beta1/nodes (kubectl proxy，该集群得部署 metrics-server deployment)
	nodeMetricsList, err := resourceMetricsClient.NodeMetricses().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err)
	}
	for _, nodeMetrics := range nodeMetricsList.Items {
		klog.Infof("nodeName:%s interval:[%s %s] usage:[cpu %s memory %s]",
			nodeMetrics.Name, nodeMetrics.Timestamp, nodeMetrics.Window,
			nodeMetrics.Usage.Cpu().String(), nodeMetrics.Usage.Memory().String())
	}

	// http://127.0.0.1:8001/apis/metrics.k8s.io/v1beta1/pods
	// metrics-server pod
	podMetricsList, err := resourceMetricsClient.PodMetricses("kube-system").List(context.TODO(), metav1.ListOptions{
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
