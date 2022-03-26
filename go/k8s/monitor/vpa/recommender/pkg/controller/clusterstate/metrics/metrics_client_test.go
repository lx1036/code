package metrics

import (
	"flag"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
	"testing"

	"k8s.io/client-go/tools/clientcmd"
)

var (
	kubeconfig = flag.String("kubeconfig", "", "kubeconfig path")
)

func TestMetricsClient(test *testing.T) {
	flag.Parse()

	if len(*kubeconfig) == 0 {
		return
	}

	restConfig, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err)
	}

	metricsClient := NewMetricsClient(restConfig, corev1.NamespaceAll)
	containerMetricsSnapshots, err := metricsClient.GetContainersMetrics()
	if err != nil {
		panic(err)
	}

	for _, containerMetricsSnapshot := range containerMetricsSnapshots {
		klog.Info(containerMetricsSnapshot)
		// INFO: 这里cpu值单位是 m
		//  &{{{kube-system cilium-794fn} cilium-agent} 2021-06-14 23:17:42 +0800 CST 30s map[cpu:3 memory:138723328]}
		//  &{{{kube-system cilium-9jn5v} cilium-agent} 2021-06-14 23:17:43 +0800 CST 30s map[cpu:3 memory:123912192]}
		//  &{{{cattle-prometheus grafana-cluster-monitoring-b6779f844-7ptkw} grafana-proxy} 2021-06-14 23:17:34 +0800 CST 30s map[cpu:1 memory:66588672]}
		//  &{{{cattle-prometheus grafana-cluster-monitoring-b6779f844-7ptkw} grafana} 2021-06-14 23:17:34 +0800 CST 30s map[cpu:16 memory:49356800]}
	}
}
