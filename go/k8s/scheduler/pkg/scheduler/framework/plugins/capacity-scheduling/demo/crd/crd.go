package main

import (
	"context"
	"flag"
	"os"

	"k8s-lx1036/k8s/scheduler/pkg/scheduler/client/clientset/versioned"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
)

var (
	kubeConfigPath = flag.String("kubeconfig", "", "kubeconfig path")
)

// debug in local: go run . --kubeconfig=`echo $HOME`/.kube/config
func main() {
	flag.Parse()

	if len(*kubeConfigPath) == 0 {
		os.Exit(1)
	}

	restConfig, err := clientcmd.BuildConfigFromFlags("", *kubeConfigPath)
	if err != nil {
		panic(err)
	}
	elasticQuotaClient, err := versioned.NewForConfig(restConfig)
	if err != nil {
		panic(err)
	}

	elasticQuotas, err := elasticQuotaClient.SchedulingV1alpha1().ElasticQuotas("quota1").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err)
	}

	klog.Info(len(elasticQuotas.Items))
	for _, elasticQuota := range elasticQuotas.Items {
		klog.Info(elasticQuota.Spec)
	}
}
